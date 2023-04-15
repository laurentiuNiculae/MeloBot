#!/bin/bash

curl -X POST 'https://id.twitch.tv/oauth2/token' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'client_id=9v1qn54zofjqbpqarcq4riz0074fuq&client_secret=7ttttrdnss44w978dyp0o8rkxdeoju&code=wq9bpwasybmjwo6bjoevsjb1zxutd5&grant_type=authorization_code&redirect_uri=https://localhost:6969'
