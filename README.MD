# /!\ This project is not maintained anymore. It was an experiment that never made it into ceph-nano. /!\ 

# cn-core

`cn-core` replaces the `demo.sh` scenario of the ceph-container project for a faster bootstrap time, better test coverage and error handling.

```
cn-core init
```

See it in action:

```text
[root@37c246814969 tmp]# ./cn-core init
2019/01/14 10:57:29 init: no daemon was selected. Deploying mon, mgr, osd, rgw, dash.
2019/01/14 10:57:29 init mon: run prerequisites
2019/01/14 10:57:29 init mon: writing monitor initial keyring
2019/01/14 10:57:29 init mon: writing ceph configuration file
2019/01/14 10:57:29 init mon: generating monitor map
2019/01/14 10:57:29 init mon: populating monitor store
2019/01/14 10:57:29 init mon: running monitor
2019/01/14 10:57:29 init mgr: run prerequisites
2019/01/14 10:57:29 init mgr: fetching admin keyring
2019/01/14 10:57:29 init mgr: generating manager keyring
2019/01/14 10:57:29 init mgr: running manager
2019/01/14 10:57:29 init osd: run prerequisites
2019/01/14 10:57:29 init osd: generating osd keyring
2019/01/14 10:57:30 init osd: populating osd store
2019/01/14 10:57:31 init osd: running osd
2019/01/14 10:57:33 init rgw: run prerequisites
2019/01/14 10:57:33 init rgw: generating rgw keyring
2019/01/14 10:57:33 init rgw: running rgw on port 8000
2019/01/14 10:57:33 init rgw: creating rgw user
2019/01/14 10:57:41 init dashboard: run prerequisites
2019/01/14 10:57:41 init dashboard: running dashboard on port 5000
2019/01/14 10:57:41 init ceph: running ceph health watcher
```

Or for daemon in particular `cn-core init --daemon mon`.

Full CLI example:

```
$ cn-core init -h
Init a Ceph cluster

Usage:
  cn-core init [flags]

Examples:
cn-core init
cn-core init --daemon mon


Flags:
  -d, --daemon string            Specify which daemon to bootstrap. Valid choices are: mon, mgr, osd, rgw, dash, health.
      --rgw-port string          Specify binding port for Rados Gateway. (default "8000")
      --dash-port string         Specify binding port for Sree dashboard. (default "5000")
      --dash-exposed-ip string   Specify binding port for Sree dashboard.
  -h, --help
```
