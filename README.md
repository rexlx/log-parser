# little log parser
in progress. please help (or don't).

### build
```
git clone https://github.com/rexlx/log-parser.git
cd log-parser/
go build -o log-parser ./*.go
```

### pipe logs to the parser
```
journalctl -f -o json | tr '[:upper:]' '[:lower:]' | log-parser -scan -stalk pmie.service
```

### read in data from files
```
# all trailing args are the files you want to read in. future versions will not require the src directory and file list
# to run as fast as humanly possible set the read value higher than you think you should
log-parser -src /s/b/logs -read 20 -stalk ssh.service $(ls /s/b/logs | tr '\n' ' ')
```


### example
```
log-parser -src /s/b/logs -read 500 -show 15 $(ls /s/b/logs | tr '\n' ' ')

rsyslog.service        658001 > 64.21006613236426
pmie.service           207469 > 20.245559217106784
init.scope             54689 > 5.336746154964612
                       50079 > 4.886886040967521
cron.service           8949 > 0.8732750889717915
user@1000.service      8543 > 0.8336561722076227
systemd-logind.service 4393 > 0.4286844860714136
session-38.scope       2493 > 0.24327576229820946
ssh.service            2380 > 0.23224882241064518
session-43.scope       2375 > 0.23176090471650518
networkmanager.service 2259 > 0.22044121421245694
dnf-makecache.service  2049 > 0.19994867105857647
dbus.service           1814 > 0.17701653943399595
wpa_supplicant.service 1798 > 0.17545520281274793
crond.service          1724 > 0.16823402093947576


read 121 files and processed 1024763 records in 2.057578801 seconds
```
