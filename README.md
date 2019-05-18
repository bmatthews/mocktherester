# MockTheRester
a very simple rest api mocker

to mock an api simply add the endpoint to the yaml file and run the application.

to build the application run `docker build -t mocktherester  .`

to run the application run `docker run -p 8080:8080 mocktherester`

if you want to supply a mocks file when running mount a volume at `/config/mocks.yaml`

## Examples

### Twillio SMS

```$yaml
routes:
  - method: POST
    name: "SMS"
    path: "/Accounts/{username}/Messages.json"
    auth:
      type: BASIC
      username: foo
      password: bar
    result:
      code: 201
      data:
         test: test

```

