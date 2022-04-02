## Bash Runtime

### Goal

Build a Bash runtime based on [Apache Pulsar](https://pulsar.apache.org/docs/en/concepts-overview/).

This runtime should be able to execute a bash script to process messages got from pulsar topics and publish results to an output topic.

### Design

The procedure of Pulsar Function is like below:

![overview](./images/pulsar-functions-overview.png)

In general, we can say that a Pulsar Function will do below things:

1. listen to one or multiple topics
2. process received messages and publish them to one or multiple output topics
3. (Optional) send logs to a log topic for debug

So it's simple to implement a runtime for bash, below is the basic idea:

1. initialize a consumer to subscribe given input topics
2. use the `exec` library to execute the external bash script when receive a message and get the result
3. initialize a producer to produce result to the out topic
4. (Optional) initialize a producer to producer log to the log topic
    - implement an `io.Writer`, which send given parameter `[]byte` to log topic in its `Write` method
    - create a new `io.Writer` using `io.MultiWriter` to combine the stdout and the custom pulsar writer
    - create a new logger using the `logrus`', and set its output to the combined writer, then all logs will be print
      to the stdout and log topic both

During the implementation, I used:

- basic bash script and golang grammars
- pulsar go client sdk to create consumers and producers, and retrieve message from input topics and produce messages
  to output topic and log topic
- golang's testing package to test my codes
- implement an `io.Writer` and set output of `logrus` for sending logs to log topic
- basic Docker and k8s knowledge
- use multi-stage builds in Dockerfile to reduce image size

### Result

- [x] **Goal1**: implement a bash script to add "!" to the end of the input message  
- [x] **Goal2**: Build a BASH Runtime
  - [x] runtime should be able to customize the input and output Pulsar topic
  - [x] runtime should be able to build consumer and producer to stream the messages from and to Pulsar topics
  - [x] runtime should be able to invoke the target BASH script
  - [x] runtime should be able to run multiple instances to support parallel processing
  - [x] runtime should be able to support Log Topic
- [x] **Goal3**: Build function as image
    - [x] build docker image
    - [x] publish to docker hub
    - [x] provide StatefulSet YAML for k8s
    - [x] provide runner image to allow user build their own function image
- [x] **Goal4**: Documentation
    - [x] describe the technical solution
    - [x] and goals that not finished
    - [x] in English
- [x] **Goal5**: Project Package
    - [x] create a GitHub Repo and push all code and document
    - [x] create a README
    - [x] available tests and samples
