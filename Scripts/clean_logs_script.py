import boto3
import re 

def lambda_handler(event, context):
    """Read file from s3 on trigger."""
    s3 = boto3.client("s3")
    if event:
        records = event.get('Records', [])
        bucketname = records[0]['s3']['bucket']['name']
        filename = records[0]['s3']['object']['key']
        fileObj = s3.get_object(Bucket=bucketname, Key=filename)
        file_content = fileObj["Body"].read().decode('utf-8')
        # remove spaces
        x = ''.join(file_content.split())
        #remove new lines
        a = x.strip().replace('\n', '')
        y = a.replace('@$', '\n')
        #create new cleaned file
        new_name = filename.split("/")[1]
        create_cleaned_logs_file(y,bucketname,new_name)
    
def create_cleaned_logs_file(new_logs,source_bucket_name,source_file_name):

    encoded_string = new_logs.encode("utf-8")

    bucket_name = source_bucket_name
    file_name = source_file_name
    lambda_path = "/tmp/" + file_name
    s3_path = "cleaned_logs/" + file_name

    s3 = boto3.resource("s3")
    s3.Bucket(bucket_name).put_object(Key=s3_path, Body=encoded_string)