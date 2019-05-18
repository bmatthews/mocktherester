# mockyrestface
a very simple rest api mocker

to mock an api simply add the endpoint to the yaml file and run the application.

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

