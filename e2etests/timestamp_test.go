// Copyright 2022 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2etests

import (
	connector "database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
)

func Test_TimestampBindings_CanBeConverted(t *testing.T) {
	db, close := newDatabase()
	defer close()

	_, err := db.Exec("CREATE TABLE mytable (t TIMESTAMP)")
	require.NoError(t, err)

	// All we are doing in this test is ensuring that writing a timestamp to the
	// database does not throw an error.
	_, err = db.Exec("INSERT INTO mytable (t) VALUES (?)", time.Now())
	require.NoError(t, err)
}

func Test_TimestampBindings_CanBeCompared(t *testing.T) {
	db, close := newDatabase()
	defer close()

	_, err := db.Exec("CREATE TABLE mytable (t TIMESTAMP)")
	require.NoError(t, err)

	// We'll insert both of these timestamps and then try and filter them.
	t0 := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(1 * time.Minute)

	_, err = db.Exec("INSERT INTO mytable (t) VALUES (?)", t0)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO mytable (t) VALUES (?)", t1)
	require.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(1) FROM mytable WHERE t > ?", t0).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func newDatabase() (*connector.DB, func()) {
	provider := sql.NewDatabaseProvider(
		memory.NewDatabase("mydb"),
	)
	engine := sqle.New(analyzer.NewDefault(provider), &sqle.Config{
		IncludeRootAccount: true,
	})
	cfg := server.Config{
		Protocol: "tcp",
		Address:  "localhost:3306",
	}
	srv, err := server.NewDefaultServer(cfg, engine)
	if err != nil {
		panic(err)
	}
	go srv.Start()

	db, err := connector.Open("mysql", "root:@tcp(localhost:3306)/mydb")
	if err != nil {
		panic(err)
	}
	return db, func() { srv.Close() }
}
