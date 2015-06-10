# Newton

[Sample of cycle time data an some graphs](https://docs.google.com/spreadsheets/d/1I-Xp5A6w9SCntkwPQcyhWsrgoKMbwCvKHpt5HQEZvY4/edit?usp=sharing)

* **cycletimer**: download cycle times from tracker and output them to `stdout` in json form
* **summarize**: take cycle time json from `stdin` and ouput summaries to `stdout`
* **planner**: take cycle time summaries from `stdin` and ouput plan to `stderr` (log)
* **dumptimes**: take cycle time json from `stdin` and output times in sorted order to `stdout` in a form that is copy-pastable into a spreadsheet

## Building

go to the binary you want to build and run `go build`. You may need to `go get` some dependencies.
