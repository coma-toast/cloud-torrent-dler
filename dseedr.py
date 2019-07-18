# actually i might abandon this for Go

from requests import HTTPSession
http = HTTPSession()

# Make a request.
r = http.request('get', 'https://httpbin.org/ip')

# View response data.
r.json()
