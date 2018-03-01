# Metron throughput tests

Code in `src/github.com/jriguera/metron-throughput`.

Note: because router (doppler) app defines its libraries with path `/internal`, golang does not 
allow to import them in a different project, so we have decided to manually copy the libs in
`src/github.com/jriguera/metron-throughput/receiver/internal`.

Once this release is deployed, you have to run the script `/var/vcap/jobs/metron-test-receiver/bin/run`
on the receiver, which is a fake doppler router. This script runs the binary in a (infinite) loop, 
the program automatically exits after 30 seconds of the latest sucessfully received log (fiter by Origin)
It will keep listening forever if it does not receive logs with the proper origin. 

The sender runs next to `metron_agent`, it will generate a lot of logs by running it with
`/var/vcap/jobs/metron-test-logger/bin/run` (see `run -h` for extra options).

With the output of both commands you can see the performance of metron_agent.

Sender:
```
generator/e11d0c70-6ba7-4cb2-b5ab-fbc789feed8e:/var/vcap/jobs/metron-test-logger/bin# ./run -runtime 600
------ Starting metron-logger 18:03:01-15:01:17
*** Init, sending to 127.0.0.1:3457 with origin metron-throughput/zz
* Starting go sub ...
* Sending termination to go sub ...
* Printing info ...
  Logs sent = 599793 (+1)
  Errors = 0
  Start time = Thu, 01 Mar 2018 15:01:17 UTC
  End time = Thu, 01 Mar 2018 15:11:17 UTC
  Elapsed time = 600.000230 s
  Rate = 999.654616
*** End
```

Receiver:
```
receiver/14fdd134-3d54-46f8-a947-d82c3b325002:/var/vcap/jobs/metron-test-receiver/bin# ./run
------ Starting metron-receiver 18:03:01-14:58:35
*** Init, listening 0.0.0.0:8082 with origin: metron-throughput/zz
* Starting Doppler Router with 1000 diodes
2018/03/01 14:58:35 Starting gRPC server on 0.0.0.0:8082
2018/03/01 15:11:47 Worker 5 about to stop due to inactivity for 30.000288 s
2018/03/01 15:11:47 Worker 1 about to stop due to inactivity for 30.000882 s
2018/03/01 15:11:47 Worker 7 about to stop due to inactivity for 30.000760 s
2018/03/01 15:11:47 Worker 4 about to stop due to inactivity for 30.000754 s
2018/03/01 15:11:47 Worker 6 about to stop due to inactivity for 30.000722 s
2018/03/01 15:11:47 Worker 0 about to stop due to inactivity for 30.000531 s
2018/03/01 15:11:47 Worker 2 about to stop due to inactivity for 30.000397 s
2018/03/01 15:11:47 Worker 3 about to stop due to inactivity for 30.000255 s
2018/03/01 15:11:47 Worker 8 about to stop due to inactivity for 30.000610 s
2018/03/01 15:11:47 Worker 9 about to stop due to inactivity for 30.000409 s
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Stopping reader in worker=9
2018/03/01 15:11:47 Worker 9 about to stop due to inactivity for 30.001567 s
* Done! Showing reports ...
* Printing Doppler library info ...
* Doppler SpyHealthRegistrar: 
   ingressStreamCount :  5
* Printing info ...
* Operations done by worker 0: 59920 (0 errors) in 599.904197 s, rate=99.882615 ops/s
* Operations done by worker 1: 61005 (0 errors) in 599.772802 s, rate=101.713515 ops/s
* Operations done by worker 2: 59815 (0 errors) in 599.934414 s, rate=99.702565 ops/s
* Operations done by worker 3: 60070 (0 errors) in 599.842736 s, rate=100.142915 ops/s
* Operations done by worker 4: 60077 (0 errors) in 599.863689 s, rate=100.151086 ops/s
* Operations done by worker 5: 59584 (0 errors) in 599.792999 s, rate=99.340939 ops/s
* Operations done by worker 6: 60440 (0 errors) in 599.732450 s, rate=100.778272 ops/s
* Operations done by worker 7: 59376 (0 errors) in 599.752730 s, rate=99.000800 ops/s
* Operations done by worker 8: 58946 (0 errors) in 599.883876 s, rate=98.262351 ops/s
* Operations done by worker 9: 60560 (0 errors) in 599.974933 s, rate=100.937550 ops/s
* Totals:
  Logs processed = 599793
  Errors = 0
  Start time = Thu, 01 Mar 2018 15:01:17 UTC
  End time = Thu, 01 Mar 2018 15:11:17 UTC
  Elapsed time = 600.005186 s
  Rate = 999.646360
*** End
```

Certificates can be generated with https://github.com/square/certstrap/


Example manifest:

```
---
name: log-test
# replace with `bosh status --uuid`
director_uuid: 39ac72ef-49ab-446b-af36-2262a9938609

# Graphite nozzle release is in our springernature organization. It is uploaded
# manually to Bosh from the repository (by pointing to the yml release file).
# You need to define the config/private.yml with the access keys to be able
# to read the bucket. The access key is in pe-pass
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
  vm_type: small
  stemcell: trusty
  vm_extensions: []
  azs:
  - z3
  networks:
  - name: cf
  jobs:
  - name: metron_agent
    release: loggregator
  - name: metron-test-logger
    release: metron-test


- name: receiver
  instances: 1
  vm_type: large
  stemcell: trusty
  vm_extensions: []
  azs:
  - z3
  networks:
  - name: cf
  jobs:
  - name: metron-test-receiver
    release: metron-test


properties:
  doppler:
    disable_announce: true
    syslog_skip_cert_verify: true
    addr: 10.230.20.250
  metron_agent:
    zone: "zz"
  loggregator:
    disable_syslog_drains: true
    tls:
      ca_cert: |
        -----BEGIN CERTIFICATE-----
        MIIE2jCCAsKgAwIBAgIBATANBgkqhkiG9w0BAQsFADANMQswCQYDVQQDEwJDQTAe
        Fw0xODAyMjAxNjA5MjNaFw0yODAyMjAxNjA5MjZaMA0xCzAJBgNVBAMTAkNBMIIC
        IjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEApmtm8BkfAX+PLa1lK47xlJG/
        6vvePqXmRW0bwgyb0tZh2rMCeHJ8v1M1JP/kfkXzUuvYnXn2ZOkN0p1UUDInet0r
        tJytE7BL8eXIrWSnsiNdhFH8CWO8jOOy7MIm3RElyvqeeyPDU2cC6k3mW9p51WKn
        pb/dWJLJNYKtl8v5G77V7a6sATUQqI7N/Zp2Vq/dkQfuv7sknkX808P3BmFS7hvo
        N6P+CNgT5FcsUo0hoZxELh5uvZPlghA553mcfl50bg01m2NBwS4kEGtZH2bv7ZYe
        k49iN7jOp9aSSR8+tRN7UsPzgLJHWIAyzZdWJzEhTpgQMIaJjB/71vaHlBwRyeE0
        jZT5BLXtHbIDIFsc29yy3a12yXZC/s5lKEiaGBm8h1mUhrl9/wd4VApFz8PVUauI
        TJt4J7DBetBhJCt4BR61ZCbC+4ZIrV3pdAw/HDMryLf8ngseLbwqE0bkz//zS6+r
        sMLjL69gbSAvKqfKasJHgOQ0D9Lhi9E7cI9X1DAl0ey5t4UTUtP+5O7nTQNZRO96
        ee6Iy5XfjZd2ltN/kjWUm671tuVfZCaCdS1J1IeDwAoLdHGnQrFyeFv4sirkQKtd
        6VnD777RncSiNActhEJ/TRg5lTAPa9V/osiqystQ3MupW6eyrWdN/6IwrS4ncUrY
        PkoocKSeGt36nFsSGRsCAwEAAaNFMEMwDgYDVR0PAQH/BAQDAgEGMBIGA1UdEwEB
        /wQIMAYBAf8CAQAwHQYDVR0OBBYEFEc4ewwtVTXj9Ea++FX13Bb064agMA0GCSqG
        SIb3DQEBCwUAA4ICAQBDkaD/4KvF5XVIFMrGyXcQpykNL+PkDzEFpcjdatjkkSU+
        IUKYlviBRW26gLFtopJum/kQA4qCZ40o/9A49KswoJugSn6FNdMPp7TirziEhR8w
        MSbei6T1FfOMXIs3gnUyUEaBko7IvI2Cv8XTfL7yX2HJHUdRUflBHFLgekbAF+qT
        JxEc6IVRNugw1og7jTgbskgy8bYAqit9e6wEKyr9Zj/a6D40p18Xnjs6IaEW0mqJ
        fzd/w4qaR1JXCc+0uqCXCwQBFZ/ykDphg2tf8wBpaN2vV8wZyh5g1VRzy92/28hI
        47JOWLFIjtqoPJ0A8GXbATHGWuAYQX1qU7Re0yB3Q/OWt/MV5aveq7reusoPLA8C
        lLizjrGwh4xIsPw9PsWt2xtqxTwUo0dM5IXU2B7goXWxGVdwWPbJp4BKTb6Mikwp
        AtpmTPiSvJWnGbt7A7ij5qMK3lnveQjLxeAkGJ60VoL5YcUW6gRouEVfTL2zOk+/
        6Wvh9KEEMWcO8POxLRQHThbeVKN+cBjjmL+6rQKrolcdfCh+44aSJlv54RsJqvxE
        5rrjIoesYev8Xq/ZrtDXzH2xBFbTEZC7HiDpu/4gnRQs92o3jEW27EvocD3Iff+C
        S2mzNqrbOzkRGGlJaPurjrg9qp4WTycSFR3ZDl6EntfT5aGSGDsGUejFZ3UnpA==
        -----END CERTIFICATE-----
      metron:
        cert: |
          -----BEGIN CERTIFICATE-----
          MIIEHzCCAgegAwIBAgIQVO65sddECJkJBFjROSqtTzANBgkqhkiG9w0BAQsFADAN
          MQswCQYDVQQDEwJDQTAeFw0xODAyMjAxNjEwNDhaFw0yMDAyMjAxNjEwNDhaMBcx
          FTATBgNVBAMMDG1ldHJvbl9hZ2VudDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
          AQoCggEBAMSC3eGMXzAPIPs89+hhpwEwxYmqVwtFhnCRFa2pcT+UT98cVYuHdPet
          Ar8UlU5cuULwwhIZKPt20ARy5q7Uw7EpKprB8TL8z32NHo0vjzcbj75wY72AQCTw
          pub4fuptwVtlSlflCdrkeSHlCiDJOmFYfTWitu6y7AkCce0oFAVsfeq5zHTyplWs
          sZ8fERhwCHhZT3G3EtZW2sa1nhpyJrbAcbHdOPlYo1+rKSLoW5mnJxe7x0f7ssfB
          zL3WVuPa9iPltXNdorHwNTNxLRum+6SmSWYw1jm24XQ7hJN7Q4dAXwS+tS+LZLfM
          JWBLaNr7z9VKGsBoXUs+9nRrAOMigJUCAwEAAaNxMG8wDgYDVR0PAQH/BAQDAgO4
          MB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAdBgNVHQ4EFgQU3WxkK+/1
          nbzsRabFSulhyLMcJ+swHwYDVR0jBBgwFoAURzh7DC1VNeP0Rr74VfXcFvTrhqAw
          DQYJKoZIhvcNAQELBQADggIBAC8F/w2YsxnsqqMvRl8f00RI0EtnxG416P8SOX1R
          Mc1PepmgfPia/lCzGdhZR5NZjKSAWB3DK+6OW6U9KRToGn8mDmEz0aRf1UJnoscS
          atX0EFulORVsyphrREZVWzDcqUrVYqo5Z1bRq4oQlZivdlmrAwFzkFYZU22KyyhW
          IlIUQHNANFsw+APDQwO6+RL9DtvX2L0BZgxyOJ4t0ghXeWnHYDVsOSVooLRKsn2M
          7N8aazdU6uGaFL25U0lp4c8gQ+ODbkiTgHI7o8a/5OI/FwJblxF8A/IyGu9v2QRD
          BLf3F4UYI1Ey0oOAzWGxldiqb8Spw6eJUbpxNIj8KRvTW114rk+GDWbe/4fV7E0L
          3WeP/VF/y5/99HC/22NkvgsvcBKMpyWed2fcI+gqPJsfxAvqg6LffhL5voEAWw/H
          V33sWLfJCmttQQFACesb1A2ch0GVsp0Zpo0Tb5I5sMkat9P4sBAuOcFe08jkVjuq
          tzOZvg8s0RDUfUX+CmLcdihxZp8LQPKknC7CxP3BucYHUnNuJYD1ukN9moH1VAXW
          3LLfd65MPWgbd4UxQditwTKnKdpazUvXdSZSVS19fBTvEmJRbVsS8eED4krBDdPt
          E9aPxdFWce7/RJaUZNvx7VRBQLt9kAg9wlg5fDtiLuaa8XuSLgID+VyfTS2h0/Ry
          Bib2
          -----END CERTIFICATE-----
        key: |
          -----BEGIN RSA PRIVATE KEY-----
          MIIEpQIBAAKCAQEAxILd4YxfMA8g+zz36GGnATDFiapXC0WGcJEVralxP5RP3xxV
          i4d0960CvxSVTly5QvDCEhko+3bQBHLmrtTDsSkqmsHxMvzPfY0ejS+PNxuPvnBj
          vYBAJPCm5vh+6m3BW2VKV+UJ2uR5IeUKIMk6YVh9NaK27rLsCQJx7SgUBWx96rnM
          dPKmVayxnx8RGHAIeFlPcbcS1lbaxrWeGnImtsBxsd04+VijX6spIuhbmacnF7vH
          R/uyx8HMvdZW49r2I+W1c12isfA1M3EtG6b7pKZJZjDWObbhdDuEk3tDh0BfBL61
          L4tkt8wlYEto2vvP1UoawGhdSz72dGsA4yKAlQIDAQABAoIBAQCtbES+YXA77I4R
          yxuJpGyLO2yJcp/A3cmonBHCoe/EyXG3l7zTF2cdkT0EPvkJIAGLVwgeir/FNHSe
          CH0Wu8Q8G/VygEgJ1FyVE65rsRY17wfrbCpJud5h+1OAMLtozhW/P+PdL8+DsvBH
          /mbyykPQVxSg+glxHMv7o4HBZwZMRYZZMPzjDnieHZHaV9d3tWUnMSNp4qo5hRoz
          pYYcrX0MyxMEC8CO56Sn6MOfGaeopTS9MOqdJMebZik4ndPi1z5NVMZp329Feawu
          lH1t9AYxv5k56wm9W1syrg/Iv1VJxLGlZSakTRIAMyvJvxdeKmqKTD92k/KL24dY
          p6RbU989AoGBAPVohTbKI8t00bfLwjajYt/YJC4o36LLI7ErN52W7J7rXq7eD0w0
          uGLbN5Kg483kqbI5SKYrSoTM1Q2D3jXlB5UljhgCmsAe7LneP6DO8lLhKYAmurvO
          Fnfe9F90gHtkyB3i/GhzM94e36VmROVrC0jCkx1k4TbSSo+CZKQu4pZfAoGBAMz+
          FvkJM0Ua9k23Zi0DTD8T1k7/oZ5eLU7PnL9oTDCXjfiZ418pQEVn6wyxfOdj1e/6
          eUYY79oLdBAhTlqmB81zjY8CEhrpaNOIf9yAQYqHbuj1pfhRqFCy3Jg9Q2zJHhKR
          SIHjadzAteFYZKwSyBdDaKIVwECnjcHENqFQtQWLAoGBAOIRSsZSJ+9AygCaP2q9
          0FOMdKfhF0KMB3Ep8q3FXmx3Pl2wSj9VQZYvg14bwD7nKjv38SjCMH9tgcZVd9oG
          BZorYl5T5+Kbmk8OoWatvSUELorTIqnnC2OZi1xzofgJux9s/j/qABnaLwPa1hTR
          Ky/3rjYhvCYYSn8xCy0D08/ZAoGBAIbv6ydbSwh+Swu1YejXduU+pZ+y3ixlSeXK
          /B9zBFQoLygqBGWrvcbyNONSIioeqcEiW5os6BXb3DaR9gXtrM0s903fyxMz+fDk
          tWXsdzg9FmD68pmXBvi4BEWibjO537XRNK1riU/q+s6vZPVwF45YrROkxbzJjqKy
          ClP90GspAoGAA6P3L+AROPx8bwwLdkHLBO565wJU9yKM0UWzcP3idoqD0/+dZ9id
          IZW+4BL5Cd6RxXnzHHiJ/hRt+sYoaegMBKffVrMWg6tI1ncwA5+HxO+zFrRpuoHF
          zoH0RovrycA1+h4CLWDXRMCbqSzfRKzlOCREj6S1EoJK4QW0butAAsI=
          -----END RSA PRIVATE KEY-----
      doppler:
        cert: |
          -----BEGIN CERTIFICATE-----
          MIIEGjCCAgKgAwIBAgIQBu5JJ+S0qmTR8D6hXhDrVDANBgkqhkiG9w0BAQsFADAN
          MQswCQYDVQQDEwJDQTAeFw0xODAyMjMxMTQwNDdaFw0yMDAyMjMxMTQwNDdaMBIx
          EDAOBgNVBAMTB2RvcHBsZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
          AQDIDl67boDn58IkaaPFhS7BQg9ZtNg/6PUujdvc5O5SshFOm2+i4iCEDphGk0cS
          1oT/oGL9KUg6mLYVyWbIb5aO4Phb9LWB0HKOvHnohLl6P8bC5ruNI0psMiBqKGCj
          SYpdEkzcG6VP9v9EdwDz7rlRS/yXDinIktIGEpLqdwGXP9OoBcbDfKwlJUcX17+z
          +xhDSPu4YTLg/qAO6ziqG+rV2nwB7Q7KNUlfN/dCcJFaJQSZOven3uojcU5wrU/Z
          7dJQ6JvEV1LRLeWkvPUFw5rASREyewlwrNQg4U+hHPukTyZjVEiKQTb7IywyJZpt
          qcHi8emeOATmlg3PyakEf2rhAgMBAAGjcTBvMA4GA1UdDwEB/wQEAwIDuDAdBgNV
          HSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFGLCS1luiK7D0wL0
          qeeq6kzR5KfCMB8GA1UdIwQYMBaAFEc4ewwtVTXj9Ea++FX13Bb064agMA0GCSqG
          SIb3DQEBCwUAA4ICAQBtOGSUxfhAHURX5OeukcsrVJbxAYos/a/H5Pwlyfrsnsh3
          pIk5fkeBzD0xHpY5lwe+1p7+KyGiWMMEO1pkG3/OEsOuep8S+35Ber/qsHUzZvzR
          B3/COQd/5Tl0DHMABVP7I0p8pjzLbFoGqyqZ/+4ImM64weCe17NXR+R8zee4WZXy
          6ZbgaRJdAljSnBs5lBLAcn75xVDgRJJlWbqEsiVEmjm+Sxj8Az4WSlQ890iQgQl0
          FYRvxrq66Kk6qpaDcl3vMLkUx/sZD8JhZ95ISLVLdNwfsw4vT26d7vE8d4N4ahnx
          Tqbo9yFPEoBCvXnYB8jXdlPwllmApdFCkvjKJHgGG583KSXTrfYKFTD+pTbzrFWv
          xZrnuegTfLV3ijrT0YEBMI/bMiCk1P2TJryhUpsBgw2iysydvbStaZHDX8ol2f5w
          UwRK1HewU4E+TntMsqRM4RQWWx7dj+Ru4JpcAZW1eW7PzJvJE9JsU1qPfwGRSOT3
          r7jp06/FmKI8V6ztOWwh+dU093jglFPBmKmPS02JjBc6PViep+WJuqGdnYUJDZsG
          ktH8OCXV5OU8k8T3Nw/WiCWbyGDMorLiGjs9vQtztGItyFbmHthR4mhINupNTKCk
          nxHnBxbOhxuuta9uxA2wkCwvA+P1TK0TlN12viyeCLSRWD9TABhLU7LRVeNxbQ==
          -----END CERTIFICATE-----
        key: |
          -----BEGIN RSA PRIVATE KEY-----
          MIIEowIBAAKCAQEAyA5eu26A5+fCJGmjxYUuwUIPWbTYP+j1Lo3b3OTuUrIRTptv
          ouIghA6YRpNHEtaE/6Bi/SlIOpi2FclmyG+WjuD4W/S1gdByjrx56IS5ej/Gwua7
          jSNKbDIgaihgo0mKXRJM3BulT/b/RHcA8+65UUv8lw4pyJLSBhKS6ncBlz/TqAXG
          w3ysJSVHF9e/s/sYQ0j7uGEy4P6gDus4qhvq1dp8Ae0OyjVJXzf3QnCRWiUEmTr3
          p97qI3FOcK1P2e3SUOibxFdS0S3lpLz1BcOawEkRMnsJcKzUIOFPoRz7pE8mY1RI
          ikE2+yMsMiWabanB4vHpnjgE5pYNz8mpBH9q4QIDAQABAoIBAEGWZHxymBRvqPij
          IawqI8/8RmgUoCkjyO5AV+qtq2y1MHNjBlCSbjKdTlMlCdIlPmlIPevd0u5TDq9J
          3kasPuIM45/SNIegvU4KgLU4fk6UBifz2V1GSqn6LSJgpn4iKBinXUd0UNhMlBfw
          JAHVLDB5BxDG9e/qIq0W/c+cwIrDL8g0AgyBntjcTPrkqtUlfdQSzGLwLVd7mZwI
          UrcpDEhmZjni1ecKV4HXE2pQqFkkDGRQivB+nIjyDMovbz388rMk4puXBUFvfrTV
          0UsKESLt4j9ZAJ5mDnyex2S6NATnCm9psZwlJ4Hmu7V6hTPg5LD2Brzj5KOo38hq
          HXQ++jkCgYEA5sh7gcL9tTSy3a0Jdt0+haPwdn3m43nObDm4R0WjAPPGyIDz5O+b
          rU5k+YGrRJASF7vDMpN9bm+DJ0qP7H1cZrdFgVzZOm0f80QL1Nzxc4oH7WtunFUa
          DXr7Yyd+oarWzG5Uqo28QeJrGQ8p8smHCeGgqp/pOryljFPI4JmaKk8CgYEA3epi
          kQ85/GoPEZFLapVkcJNH1ZWkgkJcTD0SjL4xboxgyDVNVOpM/cnIFhD61j15rvnE
          INu+VZrUtMnYK5uF8zyEOR1jKGwx+apj5I0+9NkwLtfz28FMnIiGcFpuOxomyn+K
          oEakzDdolZOfSKDYlBiK5KUbRqeAtF5od2q5O88CgYAnkCj0Jtxdiyo6rGZZ9TW5
          rVAU0CKbzo7fqMl5lmuKR0BFsS2eiqEShcTzrRIST+x6GxsseXJgU0eVnceskBUe
          Gr8UnTk0Ne7rQjgRBstxtjEDt44fyMsNko60AdpIlsP6CdQD5QZn+QvJIPtc/sVi
          oUZs2bse8aYjt11Re6OdKwKBgCbdpAGv3wH8OUNkZQb3vy2QPeaEXNmLccrQb21C
          6jloUJL/8tlKZ82TB34F30iiX6trhxQSKFWp1lMLfta0WFNvZ+Dw6qrruBz34KLo
          sfwEBdJOdCEqy5YmuxT2YZPsUprol4jWlopFsgVwY1c/BG97lOfSmuJW982fM0Cm
          6mY1AoGBAL67w5Zxh20wiwjEnC45whOUr4sytYPCjN/zPkCecTx/73h8JB6oQU8r
          VglDT27hR1jv6L93yu9wBuzCWk4+bhaXVrVGfMNdCM8vFJeCR6lP6gXVrkEf8MU0
          927W161lAhLBORq4IBd3ZPybk6URJk7fZ6M2b8aNK66R+Re/QaeW
          -----END RSA PRIVATE KEY-----


update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000
```


# Development

After cloning this repository, run:

```
./bosh_prepare

bosh create release --force && bosh upload release && bosh -n deploy
```


# License


Jose Riguera, SpringerNature Platform Engineering

Copyright 2017 Springer Nature

Apache 2.0 License

