from docker import APIClient
import json
from io import BytesIO
from dockerfile_parse import DockerfileParser
from pprint import pprint

def to_txt(stream):
    obj = stream.decode('utf8')
    return json.loads(obj)

dockerfile = '''
FROM ubuntu:latest
RUN apt-get update && apt-get install -y \
    vim
CMD ["docker", "version", "--format", "'{{json .Client.Version}}'"]
'''
dfp = DockerfileParser()
dfp.content = dockerfile
pprint(dfp.structure)
f = BytesIO(dockerfile.encode('utf-8'))
cli = APIClient()
response = [print(to_txt(line)) for line in cli.build(
    fileobj=f, rm=True, tag='yourname/volume'
)]

