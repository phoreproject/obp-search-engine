Banserver and Spam filtering
=============================

This server is created to ban some content from Phore's search engines manually and automatically.


Installation
------------

`npm run install`

Running the server
------------------

`PASSWORD=test PORT=8009 RDS_HOSTNAME=localhost RDS_USERNAME=#username# RDS_PASSWORD=#password# RDS_DB_NAME=obpsearch RDS_PORT=3306 npm run start`

And then you can access the interface using basic auth: 
`phoreadmin` as username and the password passed on the command as password (on the example: `test`)