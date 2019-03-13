
v0.3.3.2 / 2019-03-13
=====================

  * apb: Use env variables to compoes the image name
  * Include full link to the github issue in commits
  * Add missing permissions to ovirt's user
  * Update Creating-an-oVirt-user-how-to.md
  * Update Creating-an-oVirt-user-how-to.md
  * Update vars.yaml
  * Update README.md (#111)
  * Creating kubernetes deployment instruction
  * Update bug_report.md
  * docs: Fix links
  * Query the api server to extract node info
  * Delete a mistakenly added Dockerfile
  * Generate code coverage report

v0.3.2.2 / 2019-02-11
=====================

  * Release tech preview
  * WIP to fix the installation using a containerZ
  * Moving docs from wiki to repo
  * Create README.md
  * flex: Make sure the entrypoint.sh removes the old directory
  * make: run tests prior to binary creation
  * flex: identify the vm by the system uuid
  * ansible: simplify playbooks
  * ansible: setup_dns uses the inventory host's domain name
  * ansible: Move more examples into vars.yaml
  * ansible: Remove redundant playbook step done by openshift_role already
  * ansible: Cleanup redundant ansible vars
  * Update customization.yaml
  * Update customization.yaml
  * ci: Remove wrong lables of the router and registry selector
  * ci: fix usage of ansible inventory hostanem
  * ci: fix setup_dns
  * Update test-flex-pod-all-in-one.yaml
  * Update install.sh
  * Add version release info for ovirt-flexvolume-driver-apb
  * Introduce the ovirt-openshift-installer
  * Use openshift/origin-ansible with openshift_ovirt role
  * ovirt-api: omit empty values for the create disk call
  * Proper thin/thick provisioing behaviour
  * automation: disable again the service catalog installation
  * vendor folder cleanup and updates after ginkgo addtion
  * Use go vet, failed the build in case of errors
  * volume-provisioner: use "ovirt-volume-provisioner" in the metadata
  * code format and import fixes
  * code format and import fixes
  * Upgrade containers to centso 7.6 golang 1.11
  * Complete the romoval of Attach method
  * ci: Prepare for openshift_ovirt role
  * ovirt-api: Remove unused Attach method
  * ovirt-api: Properly handle various return with no body
  * ovirt-mini-api: Handle 302 error properly
  * cloud-provider: entrypoint invokes the cloud controller command first
  * cloud-provider: Set the name to ovirt-cloud-provider
  * cloud-provider: return InstanceNotFound for InstanceID when VM is down
  * cloud-provier: move private methods to the eof
  * cloud-provider: Implement ExternalID
  * cloud-provider: Implement NodeAddressesByProviderID
  * Fix type in README.md
  * cloud-provider: implement few Instances method
  * cloud-provider: Api structure changes plus tests
  * cloud-provider: Use the client reference instead of value
  * cloud provider fixes WIP
  * dep: Remove undeeded loging dependency
  * flex: Excplicitly specify the destination of the conf file
  * ci: adjust to okd 3.11
  * Update integ.ini
  * Best container practices
  * ci: Prepare for 3.11, by default install 3.11
  * make: generify targets and cleanups
  * Produce a container list after pushing them
  * Revert "ci: Install okd 3.11"
  * ci: Install okd 3.11
  * ci: refresh the container with ovirt-ansible-vm-inrfa 1.1.11
  * Update issue templates
  * Update README.md
  * Update README.md

v0.3.2.2 / 2019-02-11
=====================



v0.3.2.1 / 2018-10-15
=====================

  * flex: Fix entrypoint
  * Build containers only, no RPMs
  * Fix the broken apb invocation
  * Publish continaers to quay.io (dump dockerhub)
  * cloud-provider: Move from rgolangh/ovirt-k8-cloud-provider
  * Describe the whole project in the README

