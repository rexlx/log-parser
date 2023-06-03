# little log parser
in progress

```
go build -o log-parser ./main.go
```

## pipe logs to the parser
```
journalctl -f -o json | tr '[:upper:]' '[:lower:]' | log-parser -scan -stalk pmie.service
```

## read in data from files
```
# all trailing args are the files you want to read in. future versions will not require the src directory and file list
# to run as fast as humanly possible set the read value higher than you think you should
log-parser -src /s/b/logs -read 20 -stalk ssh.service $(ls /s/b/logs | tr '\n' ' ')
```