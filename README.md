# htmlstrip
Strips HTML from the input, outputs plain text. It is streamed in realtime without preloading the whole document.

* Easy to use Writer interface: \
    ```io.Copy(&htmlstrip.Writer{W: os.Stdout}, os.Stdin)```

* All it does is strip HTML into plain text.

* Should never use excessive memory as it does not buffer the whole document.

* Script, style and head tags are removed entirely, as they are not part of the page's text.

* The provided command strips HTML from standard input or specified files, writes plain text to standard output. \
    ```go install github.com/millerlogic/htmlstrip/cmd/htmlstrip```

* Could be used as an extremely basic, non-interactive text browser: \
    ```curl -s -S https://en.wikipedia.org/wiki/Chinchilla | htmlstrip | less```
