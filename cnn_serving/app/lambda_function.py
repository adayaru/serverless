import boto3
#from tensorflow import keras
#from tensorflow.python.keras.preprocessing import image
#from tensorflow.python.keras.applications.resnet50 import preprocess_input, decode_predictions


import tensorflow.keras
from tensorflow.keras.preprocessing import image

#from tensorflow.python.keras.preprocessing import image

from tensorflow.keras.applications.resnet50 import decode_predictions, preprocess_input
#from tensorflow.python.keras.applications.resnet50 import preprocess_input, decode_predictions

#print('after import image')

#print('after import preprocess_input...')



import numpy as np
import os
import json

import uuid
from time import time

from squeezenet import SqueezeNet
#print('after import SqueezeNet...')


#-----------------------------------------------------------------------------------------
# The following are needed for print_logline to work fine 
#-----------------------------------------------------------------------------------------
from datetime import datetime
# The following needed "sudo pip install pytz tzlocal" to be run on the Linux box first
from pytz import timezone
#import time
from time import time
#-----------------------------------------------------------------------------------------

#import logging
#import os
LOGLEVEL = os.environ.get('LOGLEVEL', 'INFO').upper()
print("LOGLEVEL is: " + LOGLEVEL)
'''
logging.basicConfig(format='%(asctime)s - %(message)s', level=LOGLEVEL)
logging.debug('Debug logging enabled!') # With a value of default LOGLEVEL of 'INFO' this will NOT get printed

def print_logline (message):
  logging.debug(message) 
  
#-----------------------------------------------------------------------------------------
# End of print_logline
#-----------------------------------------------------------------------------------------
'''


def time_now(in_tzStr):
  #format = "%a %d%b%Y %H:%M:%S"
  format = "%a %d%b%Y %H:%M:%S.%f"
  now_utc = datetime.now(timezone('UTC'))
  now_asia = now_utc.astimezone(timezone(in_tzStr))
  return now_asia.strftime(format)
#-----------------------------------------------------------------------------------------
# End of time_now in 
#-----------------------------------------------------------------------------------------

def print_logline (message):
  if LOGLEVEL == "DEBUG":
      #pline = time.asctime( time.localtime(time.time()) ) + ': ' + message
      tz_str = 'Asia/Kolkata'
      pline = time_now(tz_str) + ': ' + message
      print(pline)
#-----------------------------------------------------------------------------------------
# End of print_logline
#-----------------------------------------------------------------------------------------

s3_client = boto3.client('s3')
tmp = "/tmp/"

print_logline('Entered the global part...')
print_logline('About to get model file from s3...!')
#Download model file weights - one time
#model_object_key = event['model_object_key']  # example : squeezenet_weights_tf_dim_ordering_tf_kernels.h5
#model_bucket = event['model_bucket']
model_object_key = 'squeezenet_weights_tf_dim_ordering_tf_kernels.h5'
model_bucket = 'serverless-project-benchmark'

#model_path = tmp + '{}{}'.format(uuid.uuid4(), model_object_key)
#s3_client.download_file(model_bucket, model_object_key, model_path)
model_path = tmp + model_object_key
print_logline('model path is: ' + model_path)

if not os.path.isfile(model_path):
    print_logline('about to download file from s3')
    s3_client.download_file(model_bucket, model_object_key, model_path)
    print_logline('Downloaded model from S3' + ' into ' + model_path)
else:
    print_logline('Model file already available in ' + model_path)

model = SqueezeNet(weights='imagenet', no_top_weights_path = model_path)
print_logline('After model creation')

input_file_seq = 0
image_filenames = ['image.jpg','image2.jpg', 'n01498041_stingray.JPEG', 'n01518878_ostrich.JPEG']
image_objects = []

input_bucket = 'serverless-project-benchmark'
input_object_key = 'image.jpg'
input_file_seq = 'input_file_seq'

for i in range(len(image_filenames)):
  download_path = tmp + image_filenames[i]
  if not os.path.isfile(download_path):
      s3_client.download_file(input_bucket, image_filenames[i], download_path)
      print_logline('Downloaded input image from S3' + ' into ' + download_path)
  else:
      print_logline('Input file already available in ' + download_path)
  img1 = image.load_img(download_path, target_size=(227, 227))
  image_objects.append(img1)
  print_logline('after loading imagefile: ' + image_filenames[i])

print_logline('after loading all images')


def predict(img_seq):
    print_logline('inside predict')
    start = time()
    # model = SqueezeNet(weights='imagenet', no_top_weights_path = model_path)
    # print_logline('After model creation')
    # img = image.load_img(img_local_path, target_size=(227, 227))
    # print_logline('after load_img')
    print_logline('img_seq is: ' + str(img_seq))
    img = image_objects[img_seq]
    x = image.img_to_array(img)
    print_logline('after img.img_to_array')
    x = np.expand_dims(x, axis=0)
    print_logline('after np.expand')
    x = preprocess_input(x)
    print_logline('after preprocess_input')
    preds = model.predict(x)
    print_logline('after model.predict')
    res = decode_predictions(preds)
    print_logline('after decode_predictions')
    latency = time() - start
    print_logline('returning from predict')
    return latency, res


def lambda_handler(event, context):
    print_logline('inside handler')
    #input_bucket = event['input_bucket']
    #input_object_key = event['input_object_key']
    input_bucket = event.get('input_bucket', 'serverless-project-benchmark')
    input_object_key = event.get('input_object_key', 'image.jpg')
    input_file_seq = event.get('input_file_seq', 0)

    #download_path = tmp + '{}{}'.format(uuid.uuid4(), input_object_key)
    #s3_client.download_file(input_bucket, input_object_key, download_path)
    # download_path = tmp + input_object_key
    # if not os.path.isfile(download_path):
        # s3_client.download_file(input_bucket, input_object_key, download_path)
        # print_logline('Downloaded input image from S3' + ' into ' + input_object_key)
    # else:
        # print_logline('Input file already available in ' + download_path)

        
    #latency, result = predict(download_path)
    latency, result = predict(input_file_seq)
    print_logline('After call to predict')
    print_logline('latency is: ' + str(latency))
    print_logline('result is: ')
    print(result)
    #return latency
    #_tmp_dic = {x[1]: {'N': str(x[2])} for x in result[0]}
    response_obj = {'result': result, 'latency': latency}
    print_logline('Returning to caller!')
    return json.dumps(response_obj, default=str)
