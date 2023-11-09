## Overview

This script is designed to automate the process of checking the correctness of a program on a set of input files.
It runs the program on each input file by piping the contents of the file to the program's stdin and then checks the
correctness of the output.

## Requirements:

- Python (version 3.x recommended)
- timeout utility available in the PATH (usually comes with UNIX-based OS, for Windows use WSL)

## How to Execute:

Execute the script using Python:

```bash
python3 checker.py <executable> [--time_limit <time_in_seconds>] [--max_time_limit_exceeded <max_limit>]
```

Where:

- `executable`: Path to the executable you want to run on the input files or a command to execute the program
- `time_limit`: (Optional) The time limit for each execution. Default is 60 seconds.
- `max_limit`: (Optional) The maximum number of times the time limit can be exceeded before the script stops.
  Default is 10.

## Examples:

A few examples for some common languages:

### Java

```bash
python3 checker.py "java -jar my_java.jar"
```

### C

```bash
python3 checker.py /home/algeng/uni/my_c_program
```

### Python

```bash
python3 checker.py "python3 program.py"
```

## Output Format:

The script will print the results in the following CSV format:

```csv
file,status,time,return,stderr
```

Where:

- `file`: The name of the input file
- `status`:  Status of execution (OK, Wrong, timelimit).
- `time`: Execution time in seconds.
- `return`: Return code of the program.
- `stderr`: The output of the program on stderr.