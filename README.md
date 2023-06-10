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
# follow the logs and print stats about a particular servcie
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
log-parser -src /s/b/logs -read 500 -level 4 $(ls /s/b/logs | tr '\n' ' ')

Initialized: 10 Jun 23 09:46 CDT | Runtime: 2.632620643s | Date: 10 Jun 23 09:46 CDT
____________________________________________________________________________________
rsyslog.service         1038340 > 57.3555064799338
pmie.service             388912 > 21.482601783735593
zinc.service             122945 > 6.7911982049959185
init.scope                91722 > 5.066511706524345
audit                     63738 > 3.5207400967101536
cron.service              13213 > 0.7298556418122824
kernel                    12615 > 0.6968235012080484
user@1000.service         12218 > 0.6748941369607558
systemd-logind.service     7188 > 0.3970485395706264
ssh.service                4614 > 0.25486671696979274
dnf-makecache.service      3828 > 0.21144989002175263
crond.service              2982 > 0.16471880147462545
session-68.scope           2918 > 0.16118358910226596
networkmanager.service     2758 > 0.15234555817136722
session-38.scope           2494 > 0.13776280713538427
session-80.scope           2483 > 0.137155192508885
session-43.scope           2377 > 0.13129999701716455
session-88.scope           2354 > 0.13002953007084786
wpa_supplicant.service     2324 > 0.12837239927130437
sshd.service               1974 > 0.10903920660996333
dbus.service               1952 > 0.10782397735696475
run-parts                  1741 > 0.09616882406684202
xrdp.service               1687 > 0.0931859886276637
packagekit.service         1658 > 0.0915840955214383
session-c2.scope           1566 > 0.08650222773617153

--------------  Logs Parsed: 1810358 | Error Rate: 17.348723291194336  --------------


read 233 files and processed 1810358 records in 2.632729252 seconds
done.
```
