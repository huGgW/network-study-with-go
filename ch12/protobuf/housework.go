package protobuf

import (
	"io"

	"github.com/huGgW/network-study-with-go/ch12/housework/v1"
	"google.golang.org/protobuf/proto"
)

func Load(r io.Reader) ([]*housework.Chore, error) {
    b, err := io.ReadAll(r)
    if err != nil {
        return nil, err
    }

    var chores housework.Chores
    err = proto.Unmarshal(b, &chores)

    return chores.GetChores(), err
}

func Flush(w io.Writer, chores []*housework.Chore) error {
    b, err := proto.Marshal(&housework.Chores{Chores: chores})
    if err != nil {
        return err
    }

    _, err = w.Write(b)

    return err
}
