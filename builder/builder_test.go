package builder

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/knocknote/carpenter/dialect/mysql"
)

var (
	db     *sql.DB
	schema = "test"
)

func init() {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("root@/%s", schema))
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	db.Exec("drop table if exists `build_test`")
	os.Exit(code)
}

func TestSingleCreate(t *testing.T) {
	new, err := getTables("./_test/table1.json")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"create table if not exists `build_test` (\n" +
			"	`id` int(11) unsigned not null auto_increment,\n" +
			"	`name` varchar(64) not null,\n" +
			"	`email` varchar(255) not null,\n" +
			"	`gender` tinyint(4) not null,\n" +
			"	`country` int(11) not null,\n" +
			"	`created_at` datetime not null,\n" +
			"	`deleted_at` datetime,\n" +
			"	primary key (`id`),\n" +
			"	unique key `name` (`name`),\n" +
			"	key `k1` (`deleted_at`),\n" +
			"	key `k2` (`gender`,`country`)\n" +
			") engine=InnoDB default charset=utf8",
	}
	actual, err := Build(db, nil, new[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("err: create: unexpected SQL returned.\nactual:\n%s\nexpected:\n%s\n", actual, expected)
	}
	for _, sql := range actual {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAlter(t *testing.T) {
	old, err := getTables("./_test/table1.json")
	if err != nil {
		t.Fatal(err)
	}
	new, err := getTables("./_test/table2.json")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"alter table `build_test`\n" +
			"	drop key `k2`,\n" +
			"	drop key `name`,\n" +
			"	drop `deleted_at`,\n" +
			"	add `uuid` varchar(64) not null first,\n" +
			"	add `icon` text not null after `email`,\n" +
			"	add unique key `email` (`email`),\n" +
			"	drop key `k1`,\n" +
			"	add key `k1` (`created_at`),\n" +
			"	add key `k3` (`gender`),\n" +
			"	modify `country` tinyint(4) not null",
	}
	actual, err := Build(db, old[0], new[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("err: alter: unexpected SQL returned.\nactual:\n%s\nexpected:\n%s\n", actual, expected)
	}
	for _, sql := range actual {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSingleDrop(t *testing.T) {
	old, err := getTables("./_test/table1.json")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"drop table if exists `build_test`",
	}
	actual, err := Build(db, old[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("err: drop: unexpected SQL returned.\nactual:\n%s\nexpected:\n%s\n", actual, expected)
	}
	for _, sql := range actual {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}
}

func getTables(filename string) (mysql.Tables, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	tables := mysql.Tables{}
	if err := json.Unmarshal(buf, &tables); err != nil {
		return nil, err
	}
	return tables, nil
}
