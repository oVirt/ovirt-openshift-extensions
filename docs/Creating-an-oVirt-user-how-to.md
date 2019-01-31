# Createing an oVirt user - how to

The oVirt user needed to perform various provisioning actions needs to have the 'StorageAdmin' role in ovirt.
Using the super user admin@internal is discouraged. Instead create a dedicated user, with the role above mentioned
on the designated data center.

Here is a small playbook to create a user `'theAdmin`' in ovirt and grant it the `'StorageAdmin'` permission on the data center `'Default'`:

```
- hosts: localhost
  connection: local
  gather_facts: false

  vars:
    ansible_python_interpreter: /usr/bin/python2.7
    #aaa_jdbc_prefix: /home/rgolan/deploy/pg95/bin # for dev environment

    users:
     - name: theAdmin
       authz_name: internal-authz
       password: 123
       valid_to: "2028-01-01 00:00:00Z"

    user_groups:
     - name: group1
       authz_name: internal-authz
       users:
        - TheAdmin

    permissions:
      - state: present
        user_name: theAdmin
        authz_name: internal-authz
        role: StorageAdmin
        object_type: data_center
        object_name: Default

  pre_tasks:
    - name: Login to oVirt
      ovirt_auth:
        url: https://ovirt-engine-fqdn/ovirt-engine/api
        username: admin@internal
        password: 123
        insecure: True
  roles:
    - ovirt.infra/roles/ovirt.aaa-jdbc
    - ovirt.infra/roles/ovirt.permissions
```