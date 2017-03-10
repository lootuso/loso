package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"losa/str"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	INVALID_TYPE = errors.New("invalid value type")
)

type Values map[string]interface{}

type Field struct {
	PrimaryKey bool
	Name       string
	Type       string
	Null       bool
	Default    string
	Comment    string
}

const (
	FIELD_TAG_NAME = "db"
	IGNORE_TAG     = "ignore"
	FIELD_TAG      = "field"
	PK_TAG         = "pk"
)

type FieldTag struct {
	data map[string]string
}

func (s *FieldTag) Get(name string) string {
	if v, ok := s.data[name]; ok {
		return v
	}
	return ""
}

func (s *FieldTag) Has(name string) bool {
	if _, ok := s.data[name]; ok {
		return true
	}
	return false
}

func NewTag(tag string) *FieldTag {
	values := strings.Split(tag, ";")
	t := &FieldTag{
		data: make(map[string]string),
	}
	for _, val := range values {
		v := strings.Split(val, ":")
		k := strings.TrimSpace(v[0])
		if len(v) > 1 {
			t.data[k] = strings.TrimSpace(v[1])
		} else {
			t.data[k] = ""
		}
	}
	return t
}

type Parser struct {
	err error
}

func (p *Parser) TableName(t reflect.Type) string {
	return strings.ToLower(t.Name())
}

func (p *Parser) ScanPk(value reflect.Value) (string, error) {
	refType := value.Type()
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		tag := NewTag(field.Tag.Get(FIELD_TAG_NAME))
		if tag.Has(PK_TAG) {
			if tag.Has(FIELD_TAG) {
				return tag.Get(FIELD_TAG), nil
			}
			return strings.ToLower(field.Name), nil
		}
	}
	return "", errors.New("not scan primary key")
}

func (p *Parser) FieldName(f reflect.StructField) (string, error) {
	tag := NewTag(f.Tag.Get(FIELD_TAG_NAME))
	if tag.Has(IGNORE_TAG) {
		return "", errors.New(f.Name + " ignored")
	}
	if tag.Has(FIELD_TAG) {
		return tag.Get(FIELD_TAG), nil
	}
	return strings.ToLower(f.Name), nil
}

func (p *Parser) Encode(v interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	value := reflect.Indirect(reflect.ValueOf(v))
	if value.Kind() != reflect.Struct {
		return nil, errors.New("needs a pointer to a struct")
	}
	refType := value.Type()
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		if name, err := p.FieldName(field); err == nil {
			result[name] = value.FieldByName(field.Name).Interface()
		}
	}
	return result, nil
}

func (p *Parser) Decode(data map[string][]byte, obj interface{}) error {
	ref := reflect.Indirect(reflect.ValueOf(obj))
	if ref.Kind() != reflect.Struct {
		return errors.New("needs a pointer to a struct")
	}
	refType := ref.Type()

	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		value := ref.Field(i)
		if name, err := p.FieldName(field); err == nil {
			var v interface{}
			if val, ok := data[name]; ok {
				switch field.Type.Kind() {
				case reflect.String:
					v = string(val)
				case reflect.Bool:
					v = string(val) == "1"
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
					if m, err := strconv.Atoi(string(val)); err == nil {
						v = m
					} else {
						return INVALID_TYPE
					}
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if m, err := strconv.ParseUint(string(val), 10, 64); err == nil {
						v = m
					} else {
						return INVALID_TYPE
					}
				case reflect.Int64:
					if m, err := strconv.ParseInt(string(val), 10, 64); err == nil {
						v = m
					} else {
						return INVALID_TYPE
					}
				case reflect.Float32, reflect.Float64:
					if m, err := strconv.ParseFloat(string(val), 64); err == nil {
						v = m
					} else {
						return INVALID_TYPE
					}
				case reflect.Struct:
					if value.Type().String() == "time.Time" {
						if timestrap, err := strconv.ParseInt(string(val), 10, 64); err == nil {
							v = time.Unix(timestrap, 0)
						} else if m, err := time.Parse("2006-01-02 15:04:05", string(val)); err == nil {
							v = m
						} else if m, err := time.Parse("2006-01-02 15:04:05.000 -0700", string(val)); err == nil {
							v = m
						} else {
							return errors.New("unsupport type:" + value.Type().String())
						}
					}
				default:
					return errors.New("unsupport type:" + field.Type.Kind().String())

				}
				value.Set(reflect.ValueOf(v))
			}
		}
	}
	return nil
}

type Model struct {
	db    *sql.DB
	parse *Parser

	//
	primaryKey string
	table      string
	distinct   string
	fields     string
	join       string
	condition  string
	groupBy    string
	orderBy    string
	having     string
	limit      string
	params     []interface{}

	//
	lastSql      string
	lastInsertId int64
	affectedRows int64
}

func (m *Model) Select(str string) *Model {
	m.fields = str
	return m
}

func (m *Model) Table(str string) *Model {
	return m.From(str)
}

func (m *Model) From(str string) *Model {
	m.table = str
	return m
}

func (m *Model) Distinct(str string) *Model {
	m.distinct = str
	return m
}

func (m *Model) Join(join, table, condition string) *Model {
	split := ""
	if m.join != "" {
		split = " , "
	}
	m.join = m.join + fmt.Sprintf("%s %v JOIN %v ON %v", split, join, table, condition)
	return m
}

func (m *Model) Where(str string, args ...interface{}) *Model {
	m.condition = fmt.Sprintf(" WHERE %v", str)
	m.params = args
	return m
}

func (m *Model) OrderBy(str string) *Model {
	split := ""
	if m.orderBy != "" {
		split = " , "
	}
	m.orderBy = m.orderBy + fmt.Sprintf(" %s ORDER BY  %v", split, str)
	return m
}

func (m *Model) GroupBy(str string) *Model {
	m.groupBy = m.groupBy + fmt.Sprintf(" GROUP BY  %v", str)
	return m
}

func (m *Model) Having(str string) *Model {
	m.having = fmt.Sprintf(" HAVING %v", str)
	return m
}

func (m *Model) Limit(args ...int) *Model {
	if len(args) > 1 {
		m.limit = fmt.Sprintf(" LIMIT %v,%v", args[0], args[1])
	} else {
		m.limit = fmt.Sprintf(" LIMIT %v", args[0])
	}
	return m
}

func (m *Model) Save(v interface{}) (interface{}, error) {
	return nil, nil
}

func (m *Model) Insert(v interface{}) (int64, error) {
	refValue := reflect.Indirect(reflect.ValueOf(v))
	if refValue.Kind() != reflect.Struct {
		return 0, errors.New("needs a pointer to a struct")
	}
	columns, err := m.parse.Encode(v)
	if err != nil {
		return 0, err
	}
	if m.table == "" {
		m.table = m.parse.TableName(reflect.TypeOf(v))
	}
	keys, values := m.parseColumns(columns)
	s := populateSql("INSERT INTO %TABLE% SET %VALUES% ", map[string]string{
		"%TABLE%":  m.table,
		"%VALUES%": keys,
	})
	return m.Execute(s, values...)
}

func (m *Model) Update(v interface{}) (int64, error) {
	refValue := reflect.Indirect(reflect.ValueOf(v))
	if refValue.Kind() != reflect.Struct {
		return 0, errors.New("needs a pointer to a struct")
	}
	columns, err := m.parse.Encode(v)
	if err != nil {
		return 0, err
	}
	if pk, err := m.parse.ScanPk(refValue); err != nil {
		return 0, err
	} else {
		if pkv, ok := columns[pk]; ok {
			m.Where(pk+" = ?", pkv)
			delete(columns, pk)
		}
	}
	if m.table == "" {
		m.table = m.parse.TableName(reflect.TypeOf(v))
	}
	keys, values := m.parseColumns(columns)
	s := populateSql("UPDATE %TABLE% SET %VALUES% %WHERE%%ORDER%%LIMIT%", map[string]string{
		"%TABLE%":  m.table,
		"%VALUES%": keys,
		"%WHERE%":  m.condition,
		"%ORDER%":  m.orderBy,
		"%LIMIT%":  m.limit,
	})
	m.params = append(values, m.params...)
	return m.Execute(s, m.params...)
}

func (m *Model) Delete(v interface{}) (int64, error) {
	if m.condition == "" {
		refValue := reflect.Indirect(reflect.ValueOf(v))
		if refValue.Kind() != reflect.Struct {
			return 0, errors.New("needs a pointer to a struct")
		}
		columns, err := m.parse.Encode(v)
		if err != nil {
			return 0, err
		}
		if pk, err := m.parse.ScanPk(refValue); err != nil {
			return 0, err
		} else {
			if pkv, ok := columns[pk]; ok {
				m.Where(pk+" = ?", pkv)
			}
		}
		if m.table == "" {
			m.table = m.parse.TableName(reflect.TypeOf(v))
		}
	}
	s := populateSql("DELETE FROM %TABLE% %WHERE%%ORDER%%LIMIT%", map[string]string{
		"%TABLE%": m.table,
		"%WHERE%": m.condition,
		"%ORDER%": m.orderBy,
		"%LIMIT%": m.limit,
	})
	return m.Execute(s, m.params...)
}

func (m *Model) QueryOne() (map[string][]byte, error) {
	m.Limit(1)
	if values, err := m.QueryAll(); err == nil {
		if len(values) > 0 {
			return values[0], nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func (m *Model) QueryAll() ([]map[string][]byte, error) {
	s := m.buildQuery()
	return m.Query(s, m.params...)
}

func (m *Model) QueryScalar() ([]byte, error) {
	value, err := m.QueryOne()
	if err == nil {
		for _, v := range value {
			return v, nil
		}
	}
	return nil, err
}

func (m *Model) FindOne(v interface{}) error {
	refValue := reflect.Indirect(reflect.ValueOf(v))
	if refValue.Kind() != reflect.Struct {
		return errors.New("needs a pointer to a struct")
	}
	if m.table == "" {
		m.table = m.parse.TableName(reflect.TypeOf(v))
	}
	if val, err := m.QueryOne(); err != nil {
		return err
	} else {
		return m.parse.Decode(val, v)
	}
}

func (m *Model) FindAll(v interface{}) error {
	refValue := reflect.Indirect(reflect.ValueOf(v))
	if refValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}
	refElem := refValue.Type().Elem()
	if refElem.Kind() != reflect.Struct {
		return errors.New("needs a pointer to a struct")
	}
	if m.table == "" {
		m.table = m.parse.TableName(refElem)
	}
	if values, err := m.QueryAll(); err != nil {
		return err
	} else {
		for _, val := range values {
			ins := reflect.New(refElem)
			if err := m.parse.Decode(val, ins.Interface()); err == nil {
				refValue.Set(reflect.Append(refValue, reflect.Indirect(reflect.ValueOf(ins.Interface()))))
			} else {
				log.Println(err)
			}
		}
	}
	return nil
}

func (m *Model) parseColumns(columns map[string]interface{}) (string, []interface{}) {
	s := ""
	params := make([]interface{}, 0)
	if len(columns) > 0 {
		for k, v := range columns {
			s = s + k + " = ?,"
			params = append(params, v)
		}
		s = strings.TrimRight(s, ",")
	}
	return s, params
}

func (m *Model) buildQuery() string {
	//'SELECT%DISTINCT% %FIELD% FROM %TABLE%%JOIN%%WHERE%%GROUP%%HAVING%%ORDER%%LIMIT% %UNION%%COMMENT%';
	s := "SELECT %DISTINCT% %FIELD% FROM %TABLE%%JOIN%%WHERE%%GROUP%%HAVING%%ORDER%%LIMIT%"
	replaceMap := map[string]string{
		"%TABLE%":    m.table,
		"%DISTINCT%": m.distinct,
		"%FIELD%":    m.fields,
		"%JOIN%":     m.join,
		"%WHERE%":    m.condition,
		"%GROUP%":    m.groupBy,
		"%HAVING%":   m.having,
		"%ORDER%":    m.orderBy,
		"%LIMIT%":    m.limit,
	}
	return populateSql(s, replaceMap)
}

func (m *Model) Query(str string, args ...interface{}) ([]map[string][]byte, error) {
	defer m.flush()
	m.lastSql = str
	stmt, err := m.db.Prepare(str)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	log.Println(m.lastSql)
	res, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	columns, err := res.Columns()
	if err != nil {
		return nil, err
	}
	values := make([][]byte, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	result := make([]map[string][]byte, 0)
	for res.Next() {
		err = res.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		value := make(map[string][]byte)
		for i, col := range values {
			value[columns[i]] = col
		}
		result = append(result, value)
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	return result, nil
}

func (m *Model) Execute(str string, args ...interface{}) (int64, error) {
	defer m.flush()
	m.lastSql = str
	log.Println(m.lastSql)
	if res, err := m.db.Exec(str, args...); err != nil {
		return 0, err
	} else {
		if id, err := res.LastInsertId(); err == nil {
			m.lastInsertId = id
		}
		if rows, err := res.RowsAffected(); err == nil {
			m.affectedRows = rows
		}
		if m.lastInsertId > 0 {
			return m.lastInsertId, nil
		} else {
			return m.affectedRows, nil
		}
	}
}

func (m *Model) LastSql() string {
	return m.lastSql
}

func (m *Model) LastInsertId() int64 {
	return m.lastInsertId
}

func (m *Model) flush() {
	m.primaryKey = ""
	m.table = ""
	m.distinct = ""
	m.fields = "*"
	m.join = ""
	m.condition = ""
	m.groupBy = ""
	m.orderBy = ""
	m.having = ""
	m.limit = ""
	m.lastInsertId = 0
	m.affectedRows = 0
	m.params = make([]interface{}, 0)
}

//replace sql
func populateSql(s string, pairs map[string]string) string {
	for k, v := range pairs {
		s = strings.Replace(s, k, v, 1)
	}
	return s
}

func New(db *sql.DB) *Model {
	m := &Model{
		db:     db,
		parse:  &Parser{},
		fields: "*",
		params: make([]interface{}, 0),
	}
	return m
}
