---
name: Bug report
about: Create a report to help us improve

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Versions (please complete the following information):**
 - OS: of the nodes, master involved `cat /etc/os-release`
 - Openshift|kubernetes version `oc version` or `kubectl version`
 - oVirt version `rpm -ql ovirt-engine`

**Logs:**
 - Openshift master and node: `journalctrl --since "-2h"`
 - volume-provisioner pod: `oc logs pods/ovirt-volume-provisioner-XYZ`
 - ovirt-engine: `/var/log/ovirt-engine/engine.log`
