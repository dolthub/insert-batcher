package main

import (
	"flag"
	"fmt"
	"github.com/OneOfOne/xxhash"
	"io"
	"log"
	"os"
	"strings"

	ast "github.com/dolthub/vitess/go/vt/sqlparser"
)

var (
	in      = flag.String("test", "", "the path to a test file")
	out     = flag.String("out", "", "result output path")
	bFactor = flag.Uint("b", 500, "batching size")
)

func main() {
	flag.Parse()

	f, err := os.ReadFile(*in)
	if err != nil {
		log.Fatalf("failed to read input file: %s\n", err)
	}

	outFile, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("failed to create output file: %s\n", err)
	}

	err = batchQueries(string(f), outFile, int(*bFactor))
	if err != nil {
		log.Fatalf("failed to create output file: %s\n", err)
	}

}

func batchQueries(in string, out io.Writer, batchSize int) error {
	var err error
	b := newBatcher(batchSize)
	err = walkQueries(in, func(n ast.Statement) error {
		return b.add(n)
	})
	if err != nil {
		return err
	}
	b.flushBatches()

	_, err = out.Write([]byte(b.out.String()))
	if err != nil {
		return fmt.Errorf("failed to write to output file: %s\n", err)
	}
	return nil
}

func newBatcher(b int) *batcher {
	return &batcher{
		batchSize: b,
		out:       &strings.Builder{},
		tables:    make(map[uint64]int),
	}
}

type batcher struct {
	batchSize int

	out *strings.Builder

	tables    map[uint64]int
	templates []*ast.Insert
	batches   [][]ast.ValTuple

	last int
}

func (b *batcher) add(n ast.Statement) error {
	switch n := n.(type) {
	case *ast.Insert:
		rows, ok := n.Rows.(ast.Values)
		if !ok {
			b.writeStatement(n)
			return nil
		}
		key, err := keyInsert(n)
		if err != nil {
			return err
		}
		if _, ok := b.tables[key]; !ok {
			b.newTemplate(key, n)
		}
		id := b.tables[key]
		b.batches[id] = append(b.batches[id], rows...)
		if b.isBatchFull(id) {
			b.writeBatch(id, b.batchSize)
		}
	default:
		b.writeStatement(n)
	}
	return nil
}

func (b *batcher) newTemplate(key uint64, n *ast.Insert) {
	b.tables[key] = len(b.templates)
	b.templates = append(b.templates, n)
	b.batches = append(b.batches, nil)
}

func (b *batcher) isBatchFull(id int) bool {
	return len(b.batches[id]) > b.batchSize
}

func (b *batcher) isBatchEmpty(id int) bool {
	return len(b.batches[id]) == 0
}

func (b *batcher) writeBatch(id, size int) {
	t := b.templates[id]
	t.Rows = ast.Values(b.batches[id][:size])
	b.writeStatement(t)
	b.batches[id] = b.batches[id][size:]
	return
}

func (b *batcher) flushBatches() {
	for id := range b.batches {
		if !b.isBatchEmpty(id) {
			b.writeBatch(id, len(b.batches[id]))
		}
	}
}

func (b *batcher) writeStatement(n ast.Statement) {
	writeToBuffer(b.out, n)
}

func writeToBuffer(b *strings.Builder, s ast.Statement) {
	ast.Append(b, s)
	b.WriteString(";\n")
}

func keyInsert(n *ast.Insert) (uint64, error) {
	hash := xxhash.New64()
	if _, err := hash.Write([]byte(fmt.Sprintf("%#v,", n.Table.String()))); err != nil {
		return 0, err
	}
	if _, err := hash.Write([]byte(fmt.Sprintf("%#v,", n.Columns))); err != nil {
		return 0, err
	}
	return hash.Sum64(), nil
}

func walkQueries(q string, cb func(stmt ast.Statement) error) error {
	var stmt ast.Statement
	var parsed string
	var err error
	remainder := q

	for len(remainder) > 0 {
		var ri int
		stmt, ri, err = ast.ParseOne(remainder)
		if err == io.EOF {
			break
		} else if remainder == "\n" {
			break
		}
		if ri != 0 && ri <= len(remainder) {
			parsed = remainder[:ri]
			parsed = strings.TrimSpace(parsed)
			if strings.HasSuffix(parsed, ";") {
				parsed = parsed[:len(parsed)-1]
			}
			remainder = remainder[ri:]
		}

		if parsed == "" {
			continue
		}

		err = cb(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
