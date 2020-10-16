## UNRELEASED

FEATURES:

IMPROVEMENTS:

* plugin/docker: can specify alternate Dockerfile path [GH-539]
* plugin/docker-pull: support registry auth [GH-545]

BUG FIXES:

* cli: fix compatibility with Windows versions prior to 1809 [GH-515]
* cli: `waypoint config` no longer crashes [GH-473]
* cli: autocomplete in bash no longer crashes [GH-540]
* entrypoint: do not block child process startup on URL service connection [GH-544]
* plugin/aws-ecs: Use an existing cluster when there is only one cluster [GH-514]

## 0.1.1 (October 15, 2020)

Initial Public Release