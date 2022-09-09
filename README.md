# serverless
This repository hosts the code that was used for comparing Lambda and vHive performance for similar code.

Details:
I did a project that compared latency across two functions on two serverless frameworks:
  . vHive 
  . AWS Lambda

For vHive, I used whatever was readily available at this repo:
https://github.com/ease-lab/vhive

I had a constraint that I could not use modified code for vHive because of this issue:
https://github.com/ease-lab/vhive/issues/68

This says: 
"The queue-proxy container interacts directly with the user workload Firecracker-Containerd container. However, we create a stub lightweight user container still so that CRI Operations such as ContainerStatus, ListContainers do not require additional support. Once firecracker-containerd supports there call, we can decommission the stub."
A related discussion is at:
https://github.com/ease-lab/vhive/discussions/587

The suggested workaround there was to use containers instead of firecracker microVM's.

Since AWS Lambda too is based on firecracker microVM, I could not use containers.
So I decided to live with this restriction.

This mandated that I needed to change Lambda code for correspoding functions to be similar to
the vHive ones.

For lambda, I modified lambda functons (that were more or less identical to the vhive 
functions) at:
https://github.com/ddps-lab/serverless-faas-workbench/tree/master/aws/cpu-memory

Since the image classification part was not eaxctly similar, it showed a huge gulf in performance
between Lambda and vHive. I made the code for Lambda simpler to make it look identical to what
vHive does.

This repo hosts the code used for comparing Lambda for two CPU-bound functions:
  . chameleon
  . cnn_serving
  
The directories under "chameleon" and "cnn_serving" has a file called lambda_function.py that implements the function
for http renderer. It has other supporting files like requirements.txt (for adding libraries to python)
and a dockerfile to build the image for deployong on Lambda.

cnn_serving also has an infer.py using which one can test the function locally (as a regular function) assuming that
you have already signed into AWS on your machine.


The corresponding code for vHive is as per the original code:
chameleon:
vhive:
See server.py at:
https://github.com/ease-lab/vhive/tree/main/function-images/chameleon 

Lambda: 
See the file lambda_function.py in the directories under "chameleon"

cnn_serving:
vHive:
See server.py at:
https://github.com/ease-lab/vhive/tree/main/function-images/cnn_serving

Lambda:
See the file lambda_function.py in the directories under "cnn_serving"
