local application = import './func.libsonnet';


application(
  env='dev',
  version='v1.2.1',
  clientId='abcd',
  port=8080
)
