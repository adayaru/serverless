
+-----------------------------------------------------------------------------+
# serverless
+-----------------------------------------------------------------------------+

# Lambda function implementations of vHive sample functions, chameleon and cnn_serving
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


awsinvoke:
This is a modified version of the invoker code available at:
https://github.com/ease-lab/vhive/tree/main/examples/invoker

invoker is described in vHive quickstart guide:
https://github.com/ease-lab/vhive/blob/main/docs/quickstart_guide.md

+-----------------------------------------------------------------------------+
# awsinvoker
+-----------------------------------------------------------------------------+

awsinvoker can be invoked to invoke AWS Lambda functions.
The functions that need to be invoked will have to included in the file endpoints.json.
Sample endpoints.json:
[
	{
		"hostname": "chameleon",
		"eventing": false,
		"matchers": null
	},
  {
		"hostname": "cnn_serving",
		"eventing": false,
		"matchers": null
	}
]

This will invoke Lambda functions called chameleon and cnn_serving.

In order to run awsinvoker, first run:
go build
in the directory awsinvoker.

Sample invocation of awsinvoker:
./awsinvoker -latf 180_chameleon_cnn.csv -endpointsFile endpoints.json -time 180 >180sec_lambda_invoke_log_`date +%Y-%m-%d.%H.%M.%S`.txt

This will make awsinvoker program to make calls to Lambda functions named in the file endpoints.json (see sample above).
The functions will be invoked as threads many times within the input time duration specified in the argument: '-time'.
The output will be written into a file whose name will be generated using current date and time as shown in the command above.

You can change endpoints.json to have your function names and vary the other args accordingly.
