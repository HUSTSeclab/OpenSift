package storage

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
)

type BatchExecContextConfig struct {
	AutoCommit     bool
	AutoCommitSize int
}

type BatchExecContext interface {
	GetSentences() string
	GetArgs() []interface{}
	Clear()
	Commit() (sql.Result, error)
	AppendExec(sentence string, args ...interface{}) error
}

type batchExecContext struct {
	appDb          AppDatabaseContext
	config         *BatchExecContextConfig
	sentences      string
	args           []interface{}
	sentencesCount int
}

func (ctx *batchExecContext) GetSentences() string {
	return ctx.sentences
}

func (ctx *batchExecContext) GetArgs() []interface{} {
	return ctx.args
}

func (ctx *batchExecContext) Clear() {
	ctx.sentences = ""
	ctx.args = make([]interface{}, 0)
	ctx.sentencesCount = 0
}

func (ctx *batchExecContext) Commit() (sql.Result, error) {
	conn, err := ctx.appDb.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ret, err := conn.Exec(ctx.sentences, ctx.args...)
	ctx.Clear()

	return ret, err
}

// AppendExec appends a sentence to the batch execution context.
// Use like db.Exec()
func (ctx *batchExecContext) AppendExec(sentence string, args ...interface{}) error {
	s := sentence
	start := len(ctx.args)

	// check if the min $n is greater than or equal to 0
	// and max $n is less than or equal to len(args)
	// if not, return an error

	// find all placeholders in the sentence
	regexp := regexp.MustCompile(`\$[0-9]+`)
	matches := regexp.FindAllStringIndex(s, -1)

	toReplace := make([]int, len(matches))

	for i, match := range matches {
		// get the number of the placeholder
		nStr := s[match[0]+1 : match[1]]
		// parse the number
		n, err := strconv.Atoi(nStr)
		if err != nil {
			return fmt.Errorf("invalid placeholder $%s in sentence %s", nStr, sentence)
		}
		if n < 1 || n > len(args) {
			return fmt.Errorf("invalid placeholder $%d in sentence %s", n, sentence)
		}
		toReplace[i] = start + n
	}

	// replace all placeholders with $n in the sentence
	for i := len(matches) - 1; i >= 0; i-- {
		n := toReplace[i]
		s = s[:matches[i][0]] + "$" + strconv.Itoa(n) + s[matches[i][1]:]
	}

	ctx.sentences += sentence + ";"
	ctx.args = append(ctx.args, args...)
	ctx.sentencesCount++

	if ctx.config.AutoCommit && ctx.sentencesCount >= ctx.config.AutoCommitSize {
		_, err := ctx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
