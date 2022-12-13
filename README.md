# Mailfetcher

[![build](https://github.com/doombeaker/mailfetcher/actions/workflows/go.yml/badge.svg)](https://github.com/doombeaker/mailfetcher/blob/master/.github/workflows/go.yml)
[![release](https://github.com/doombeaker/mailfetcher/actions/workflows/release.yml/badge.svg)](https://github.com/doombeaker/mailfetcher/blob/master/.github/workflows/release.yml)

Mailfetcher can help you download attchment of mails in batch.

## How to use

### Luanch mailfetcher

Usage will be shown with `mailfetcher -h`:

```text
usage:
        mailfetcher.exe -l
                list configs and related index
        mailfetcher.exe -i [<start> <end>]
                interactive mode
        mailfetcher.exe -d <index>
                download mails using default date setting
                FROM 202212112130 TO 202212121730
        mailfetcher.exe -s <start> <end> <index>
                download mails using date setting manually
                eg: mailfetcher.exe -s 202212112130 202212121730 0
        mailfetcher.exe -h
                show this
```

The `<start>` and `<end>` string represented time limits the time range. Only the mail whose receving time is within the time range will be fetched.

Three mode are provided:

- interactive mode
- default mode
- set manually mode

**NOTE**: the time shown in the usage prompt will be updated automatically, it's convenient for one to copy and run the example in usage directly.

### Config file

There can be multiple config files whose extension is `.txt` in `configs` directory.

**NOTE**: the `.txt` file must be encoded as UTF-8 WITHOUT BOM.

Each `.txt` file has three sections which split by blank lines.

```text
KEY-VALUE Setting Section
...
<blank line>
Name List Section
...
<blank line>
Others
....
```

Take `EXAMPLE.txt` for example:

```text
homework_path=c:\homework    # 所有班级作业将存放至的根目录
mailserver=example.com:993   # IMAP邮件服务器地址
mail_user=username           # 账号
mail_passwd=password         # 密码
prefix_flag=SS1201           # 附件前缀
maxmail=40                   # 一次最多下载邮件数目
delimiter=-                  # 邮件主题&附件的分隔符

张三
李四
王五

#以下人员暂时不在
赵六
孙七
```

### Format of mail subject & attachment

The mail subject **AND** filename of attachment must be in the following format:

```text
<prefix_flag><delimiter><name><delimiter><others>
```

Take [EXAMPLE.txt](./configs/EXAMPLE.txt) as example, in which we set:

```text
prefix_flag=SS1201           # 附件前缀
delimiter=-                  # 邮件主题/附件分隔符

张三
```

So the following subject or attachment file name meet the requirements:

```text
SS1201-张三-第一天笔记
SS1201-张三-第一天笔记.zip
```

## View the data of file data.db

The indiscipline cases will be stored to `./data.db` which is a sqlite database file.

The data in `./data.db` can be viewed by sqlite client, such as [SqliteStudio](https://sqlitestudio.pl/), [sqlite-web](https://github.com/coleifer/sqlite-web)

Take sqlite-web as example, run command below to install sqlite-web:

```bash
pip install sqlite-web
```

run command below to browse `data.db`:

```bash
sqlite_web ./data.db
```