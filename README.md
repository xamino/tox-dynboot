# tox-dynboot
Go package that fetches the bootstrap nodes from the [Tox wiki](https://wiki.tox.im/Nodes) to use with a Golang wrapped Toxcore.

## Usage
The basic case is that we want any single working node.
The following function returns either the fastest responding ToxNode or an error.
To avoid locking up a program indefinitely a timeout must be given which will trigger an error if reached.

```go
toxNode, err := toxdynboot.FetchFirstAlive(100 * time.Millisecond)
```

Beyond that further methods are provided.
For a more in depth overview see the documented source code.

## License

The MIT License (MIT)

Copyright (c) 2015 Tamino Hartmann

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
