#!/usr/bin/expect
# Bash ftp download. This is silly. Just use python. 
set HOST "sftp.bitport.io"
set USER "jdale215@gmail.com"
set PASSWORD ">hh]dR4KK2%:+?n^~u%J\n"

spawn sftp -P 2022 $USER@$HOST
expect "password:"
send $PASSWORD
# spawn bash -c 
expect "sftp>"
send "get -r TV\\ Shows /media/jason/3C72B82272B7DF38/Shows/\n"
expect "sftp>"
send "get -r Movies /media/jason/3C72B82272B7DF38/Movies/\n"
expect "sftp>"
send "exit\n"
interact
