FROM docker.elastic.co/app-search/app-search:7.4.0
HEALTHCHECK --interval=1s --retries=300 --start-period=60s CMD python -c 'import urllib, json; response = urllib.urlopen("http://myelastic:changeme@localhost:3002/swiftype-app-version"); data = json.loads(response.read()); exit(0);'
