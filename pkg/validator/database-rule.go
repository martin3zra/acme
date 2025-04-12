package validator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"

	"github.com/martin3zra/acme/pkg/database"
)

type DatabaseRule struct {
	key              string
	attributeValue   reflect.Value
	attributes       []string
	wheres           []string
	db               *sql.DB
	ignoreGivenValue any
	ignoreColumn     string
}

func newDatabaseRule(ctx context.Context, key string, attributes []string, value reflect.Value) *DatabaseRule {
	newBDRule := &DatabaseRule{
		key:              key,
		attributes:       attributes,
		attributeValue:   value,
		wheres:           make([]string, 0),
		ignoreGivenValue: nil,
	}

	newBDRule.resolveConnection(ctx)
	return newBDRule
}

func (d *DatabaseRule) getCount() int {

	var count int64
	var values = []any{d.resolveValue()}
	if d.ignoreGivenValue != nil {
		values = append(values, d.ignoreGivenValue)
	}

	err := d.db.QueryRow(d.compileSqlStatement(), values...).Scan(&count)
	if err != nil {
		fmt.Printf("an error has occurred: %v\n%s\n", err, d.compileSqlStatement())
		return 0
	}

	return int(count)
}

func (d *DatabaseRule) ignore(ignore any, column string) *DatabaseRule {
	d.ignoreGivenValue = ignore
	d.ignoreColumn = column
	return d
}

func (d *DatabaseRule) resolveTableName() string {
	return d.attributes[0]
}

func (d *DatabaseRule) resolveColumnName() string {
	if len(d.attributes) == 1 {
		return d.key
	}

	return d.attributes[1]
}

func (d *DatabaseRule) resolveValue() any {
	switch d.attributeValue.Kind() {
	case reflect.Int:
		return d.attributeValue.Int()
	case reflect.String:
		return d.attributeValue.String()
	default:
		log.Fatalf("not accepted type: %v", d.attributeValue)
		return ""
	}
}

func (d *DatabaseRule) compileSqlStatement() string {

	return fmt.Sprintf(
		"select count(*) from %s where %s = $1 %s",
		d.resolveTableName(),
		d.resolveColumnName(),
		d.criteriaIfIgnoringGivenValue(),
	)
}

func (d *DatabaseRule) criteriaIfIgnoringGivenValue() string {
	if d.ignoreGivenValue != nil {
		return fmt.Sprintf("AND %s <> $2", d.ignoreColumn)
	}

	return ""
}

func (d *DatabaseRule) resolveConnection(ctx context.Context) {
	d.db = ctx.Value(database.ConnectionKey{}).(*sql.DB)

	if d.db == nil {
		panic("database connection need to be set.")
	}
}
