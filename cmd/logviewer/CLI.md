
# Logviewer App

The **Dharitri Logviewer App** exposes the following Command Line Interface:

```
$ logviewer --help

NAME:
   Dharitri Logviewer App - Logviewer application used to communicate with sme-dharitri node to log the message lines
USAGE:
   logviewer [global options]
   
AUTHOR:
   The Dharitri Team <contact@dharitri.org>
   
GLOBAL OPTIONS:
   --address value            Address and port number on which the application will try to connect to the sme-dharitri node (default: "127.0.0.1:8080")
   --level value              This flag specifies the logger levels and patterns (default: "*:INFO ")
   --file                     Will automatically log into a file
   --working-directory value  The application will store here the logs in a subfolder.
   --use-wss                  Will use wss instead of ws when creating the web socket
   --correlation              Will include log correlation elements
   --logger-name              Will include logger name
   --help, -h                 show help
   --version, -v              print the version
   

```

