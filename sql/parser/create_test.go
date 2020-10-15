package parser

import (
	"context"
	"testing"

	"github.com/genjidb/genji/database"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/sql/query"
	"github.com/stretchr/testify/require"
)

func TestParserCreateTable(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected query.Statement
		errored  bool
	}{
		{"Basic", "CREATE TABLE test", query.CreateTableStmt{TableName: "test"}, false},
		{"If not exists", "CREATE TABLE IF NOT EXISTS test", query.CreateTableStmt{TableName: "test", IfNotExists: true}, false},
		{"With primary key", "CREATE TABLE test(foo INTEGER PRIMARY KEY)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					PrimaryKeys: []int{0},
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsPrimaryKey: true},
					},
				},
			}, false},
		{"With primary key twice", "CREATE TABLE test(foo PRIMARY KEY PRIMARY KEY)",
			query.CreateTableStmt{}, true},
		{"With type", "CREATE TABLE test(foo INTEGER)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue},
					},
				},
			}, false},
		{"With not null", "CREATE TABLE test(foo NOT NULL)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), IsNotNull: true},
					},
				},
			}, false},
		{"With not null twice", "CREATE TABLE test(foo NOT NULL NOT NULL)",
			query.CreateTableStmt{}, true},
		{"With type and not null", "CREATE TABLE test(foo INTEGER NOT NULL)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsNotNull: true},
					},
				},
			}, false},
		{"With not null and primary key", "CREATE TABLE test(foo INTEGER NOT NULL PRIMARY KEY)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					PrimaryKeys: []int{0},
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsPrimaryKey: true, IsNotNull: true},
					},
				},
			}, false},
		{"With primary key and not null", "CREATE TABLE test(foo INTEGER PRIMARY KEY NOT NULL)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					PrimaryKeys: []int{0},
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsPrimaryKey: true, IsNotNull: true},
					},
				},
			}, false},
		{"With multiple constraints", "CREATE TABLE test(foo INTEGER PRIMARY KEY, bar INTEGER NOT NULL, baz[4][1].bat TEXT)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					PrimaryKeys: []int{0},
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsPrimaryKey: true},
						{Path: parsePath(t, "bar"), Type: document.IntegerValue, IsNotNull: true},
						{Path: parsePath(t, "baz[4][1].bat"), Type: document.TextValue},
					},
				},
			}, false},
		{"With multiple primary keys", "CREATE TABLE test(foo PRIMARY KEY, bar PRIMARY KEY)",
			query.CreateTableStmt{}, true},
		{"With multiple primary keys using table constraint", "CREATE TABLE test(foo integer, bar integer, PRIMARY KEY(foo, bar))",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					PrimaryKeys: []int{0, 1},
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "foo"), Type: document.IntegerValue, IsPrimaryKey: true},
						{Path: parsePath(t, "bar"), Type: document.IntegerValue, IsPrimaryKey: true},
					},
				},
			}, false},
		{"With multiple primary keys using invalid table constraint", "CREATE TABLE test(foo integer, bar integer, PRIMARY KEY(foo, not_exists))",
			query.CreateTableStmt{}, true},
		{"With all supported fixed size data types",
			"CREATE TABLE test(d double, b bool)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "d"), Type: document.DoubleValue},
						{Path: parsePath(t, "b"), Type: document.BoolValue},
					},
				},
			}, false},
		{"With all supported variable size data types",
			"CREATE TABLE test(i integer, b blob, byt bytes, t text, a array, d document)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "i"), Type: document.IntegerValue},
						{Path: parsePath(t, "b"), Type: document.BlobValue},
						{Path: parsePath(t, "byt"), Type: document.BlobValue},
						{Path: parsePath(t, "t"), Type: document.TextValue},
						{Path: parsePath(t, "a"), Type: document.ArrayValue},
						{Path: parsePath(t, "d"), Type: document.DocumentValue},
					},
				},
			}, false},
		{"With integer aliases types",
			"CREATE TABLE test(i int, ii int2, ei int8, m mediumint, s smallint, b bigint, t tinyint)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "i"), Type: document.IntegerValue},
						{Path: parsePath(t, "ii"), Type: document.IntegerValue},
						{Path: parsePath(t, "ei"), Type: document.IntegerValue},
						{Path: parsePath(t, "m"), Type: document.IntegerValue},
						{Path: parsePath(t, "s"), Type: document.IntegerValue},
						{Path: parsePath(t, "b"), Type: document.IntegerValue},
						{Path: parsePath(t, "t"), Type: document.IntegerValue},
					},
				},
			}, false},
		{"With double aliases types",
			"CREATE TABLE test(dp DOUBLE PRECISION, r real, d double)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "dp"), Type: document.DoubleValue},
						{Path: parsePath(t, "r"), Type: document.DoubleValue},
						{Path: parsePath(t, "d"), Type: document.DoubleValue},
					},
				},
			}, false},

		{"With text aliases types",
			"CREATE TABLE test(v VARCHAR(255), c CHARACTER(64), t TEXT)",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "v"), Type: document.TextValue},
						{Path: parsePath(t, "c"), Type: document.TextValue},
						{Path: parsePath(t, "t"), Type: document.TextValue},
					},
				},
			}, false},

		{"With errored text aliases types",
			"CREATE TABLE test(v VARCHAR(1 IN [1, 2, 3] AND foo > 4) )",
			query.CreateTableStmt{
				TableName: "test",
				Info: database.TableInfo{
					FieldConstraints: []database.FieldConstraint{
						{Path: parsePath(t, "v"), Type: document.TextValue},
					},
				},
			}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := ParseQuery(context.Background(), test.s)
			if test.errored {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, q.Statements, 1)
			require.EqualValues(t, test.expected, q.Statements[0])
		})
	}
}

func TestParserCreateIndex(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected query.Statement
		errored  bool
	}{
		{"Basic", "CREATE INDEX idx ON test (foo)", query.CreateIndexStmt{IndexName: "idx", TableName: "test", Path: parsePath(t, "foo")}, false},
		{"If not exists", "CREATE INDEX IF NOT EXISTS idx ON test (foo.bar[1])", query.CreateIndexStmt{IndexName: "idx", TableName: "test", Path: parsePath(t, "foo.bar[1]"), IfNotExists: true}, false},
		{"Unique", "CREATE UNIQUE INDEX IF NOT EXISTS idx ON test (foo[3].baz)", query.CreateIndexStmt{IndexName: "idx", TableName: "test", Path: parsePath(t, "foo[3].baz"), IfNotExists: true, Unique: true}, false},
		{"No fields", "CREATE INDEX idx ON test", nil, true},
		{"More than 1 path", "CREATE INDEX idx ON test (foo, bar)", nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := ParseQuery(context.Background(), test.s)
			if test.errored {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, q.Statements, 1)
			require.EqualValues(t, test.expected, q.Statements[0])
		})
	}
}
