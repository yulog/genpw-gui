// Original: https://github.com/mattn/genpw
// License: MIT
// Author: Yasuhiro Matsumoto (a.k.a. mattn)

package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
)

var (
	numbers = []byte("0123456789")
	symbols = []byte("!\"#$%&'()=~|-^\\`{*}<>?_@[;:],./'")
)

func count(bs []byte, bp []byte) int {
	n := 0
	for _, b := range bs {
		if bytes.Index(bp, []byte{b}) != -1 {
			n++
		}
	}
	return n
}

func run(w io.Writer, o, n, nc, sc int) error {
	if o <= 0 || n < 1 {
		return flag.ErrHelp
	}
	nnc := 0
	if nc > 0 {
		nnc = nc
	}
	nsc := 0
	if sc > 0 {
		nsc = sc
	}
	if n < nnc+nsc {
		return errors.New("total length of numbers+symbols is greeter than maximum characters")
	}

	var buf bytes.Buffer
	if n > nnc+nsc {
		for r := 'a'; r <= 'z'; r++ {
			buf.WriteRune(r)
		}
		for r := 'A'; r <= 'Z'; r++ {
			buf.WriteRune(r)
		}
	}
	if nc != 0 && nsc != n {
		buf.Write(numbers)
	}
	if sc != 0 && nnc != n {
		buf.Write(symbols)
	}
	chars := buf.Bytes()

	no := 0
	nw := 0
	for {
		var pw bytes.Buffer
		for {
			pw.Reset()
			for i := 0; i < n; i++ {
				r, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
				if err != nil {
					panic(err)
				}
				pw.WriteByte(chars[int(r.Int64())])
			}

			if nc > 0 && count(pw.Bytes(), numbers) < nc {
				continue
			}
			if sc > 0 && count(pw.Bytes(), symbols) < sc {
				continue
			}
			break
		}
		fmt.Fprint(w, pw.String())

		nw += n + 1
		if nw < 80-n {
			fmt.Fprint(w, " ")
		} else {
			fmt.Fprintln(w)
			nw = 0
		}
		no++
		if no == o {
			break
		}
	}
	fmt.Fprintln(w)

	return nil
}
