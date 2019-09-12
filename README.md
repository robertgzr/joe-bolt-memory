<h1 align="center">Joe Bot - Bolt Memory</h1>
<p align="center">Integration Joe with Bolt. https://github.com/go-joe/joe</p>
<p align="center">
	<a href="https://github.com/robertgzr/joe-bolt-memory/releases"><img src="https://img.shields.io/github/tag/robertgzr/joe-bolt-memory.svg?label=version&color=brightgreen"></a>
	<a href="https://godoc.org/github.com/robertgzr/joe-bolt-memory"><img src="https://img.shields.io/badge/godoc-reference-blue.svg?color=blue"></a>
</p>

---

This repository contains a module for the [Joe Bot library][joe]. Built using 
[etcd-io/bbolt][bbolt].

## Getting Started

This library is packaged using [Go modules][go-modules]. You can get it via:

```
go get github.com/robertgzr/joe-bolt-memory
```

### Example usage

```go
package main

import (
	"github.com/go-joe/joe"
	"github.com/robertgzr/joe-bolt-memory"
)

func main() {
	b := joe.New("example-bot",
		bolt.Memory(os.Getenv("DB_PATH")),
		…
	)
	
	b.Respond("remember (.+) is (.+)", b.Remember)
	b.Respond("what is (.+)", b.WhatIs)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *Bot) Remember(msg joe.Message) error {
	key, value := msg.Matches[0], msg.Matches[1]
	msg.Respond("OK, I'll remember %s is %s", key, value)
	return b.Store.Set(key, value)
}

func (b *Bot) WhatIs(msg joe.Message) error {
	key := msg.Matches[0]
	var value string
	ok, err := b.Store.Get(key, &value)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve key %q from brain", key)
	}
	if ok {
		msg.Respond("%s is %s", key, value)
	} else {
		msg.Respond("I do not remember %q", key)
	}
	return nil
}
```
## License

[BSD-3-Clause](LICENSE)

[joe]: https://github.com/go-joe/joe
[bbolt]: https://github.com/etcd-io/bbolt
[go-modules]: https://github.com/golang/go/wiki/Modules
