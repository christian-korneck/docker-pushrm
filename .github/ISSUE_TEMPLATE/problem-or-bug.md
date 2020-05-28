---
name: problem or bug
about: I have a problem with docker-pushrm
title: ''
labels: ''
assignees: ''

---

**Describe the problem or bug**
A clear and concise description of what the problem or bug is, including all steps necessary to reproduce it.

**docker-pushrm version**
check version with:
- as Docker CLI plugin: `docker --help | grep pushrm`
- as standalone: `docker-pushrm docker-cli-plugin-metadata | grep -i version`

**Docker CLI version and platform**
check with: `docker version`
if on Linux, what distro and distro version are you running? (example: Fedora 32)

**if possible: registry server version**
example: *self-hosted `harbor` version 2.0.0*

**exact command that you're running**
example: `docker pushrm --provider quay quay.io/my-user/my-repo`

**debug output**
re-run that command with `--debug` flag and paste the output here
(i.e.: `docker pushrm --provider quay quay.io/my-user/my-repo --debug`)

**Additional context**
Add any other context about the problem here.
