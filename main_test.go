package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBatchQueries(t *testing.T) {
	tests := []struct {
		name  string
		batch int
		in    string
		exp   string
	}{
		{
			name:  "simple",
			batch: 3,
			in: `
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy (y,x) values ('6',6);
`,
			exp: `insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5), ('6', 6);
`,
		},
		{
			name:  "create table",
			batch: 3,
			in: `create table xy (
    x int primary key,
    y int
);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy (y,x) values ('6',6);
`,
			exp: `create table xy (
	x int primary key,
	y int
);
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5), ('6', 6);
`,
		},
		{
			name:  "non standard insert",
			batch: 3,
			in: `create table xy (x int primary key, y int);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy select x+1, y+1 from xy where x > 0;
`,
			exp: `create table xy (
	x int primary key,
	y int
);
insert into xy select x + 1, y + 1 from xy where x > 0;
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5);
`,
		},
		{
			name:  "two tables",
			batch: 7,
			in: `create table xy (
	x int primary key,
	y int
);
create table ab (
	a int primary key,
	b int
);
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into ab values (0, '0'), (1, '1'), (2, '2');
insert into xy values (3, '2'), (4, '4'), (5, '5');
insert into ab values (3, '2'), (4, '4'), (5, '5');
insert into xy values (6, '6'), (7, '7'), (8, '8');
insert into ab values (6, '6'), (7, '7'), (8, '8');
`,
			exp: `create table xy (
	x int primary key,
	y int
);
create table ab (
	a int primary key,
	b int
);
insert into xy values (0, '0'), (1, '1'), (2, '2'), (3, '2'), (4, '4'), (5, '5'), (6, '6');
insert into ab values (0, '0'), (1, '1'), (2, '2'), (3, '2'), (4, '4'), (5, '5'), (6, '6');
insert into xy values (7, '7'), (8, '8');
insert into ab values (7, '7'), (8, '8');
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			err := batchQueries(tt.in, b, tt.batch)
			require.NoError(t, err)
			require.Equal(t, tt.exp, b.String())
		})
	}
}
