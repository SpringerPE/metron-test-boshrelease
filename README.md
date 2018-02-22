# Metron throughput tests


```
releases:
- name: loggregator
  version: 101.6
  url: https://bosh.io/d/github.com/cloudfoundry/loggregator-release?v=101.6
  sha1: f83ca26da62276e4e8247e32dbfbf2ae1a4c6138
- name: metron-test
  version: latest

stemcells:
- alias: trusty
  name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
  version: latest


instance_groups:
- name: generator
  instances: 1
  vm_type: autoscaler
  stemcell: trusty
  vm_extensions: []
  azs:
  - z3
  networks:
  - name: autoscaler
  jobs:
  - name: metron_agent
    release: loggregator
  - name: metron-tests
    release: metron-test
  properties:
    metron_agent:
      deployment: "log-test-deployment"
      zone: "z3"
    loggregator: ....
```


# Development

After cloning this repository, run:

```
./bosh_prepare
```

Springer Nature Platform Engineering

Copyright 2017 Springer Nature



# License

Apache 2.0 License

