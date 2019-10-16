# Clavis - The Key

Clavis is a simple binary for auto-provisioning the initial user for Rstudio-Connect. While there is a large array of paramters that can be provided, only three are required to operate:

## Parameters

* `email` : The email address used for creating the account
* `name` : The full name of the user used in creating the account
* `organization` : The name of the organization to use in the .motd file displayed at login
* `username` : The username to be provisioned in RSConnect. If not provided, this will default to the unix shell user issuing the command
* `file` : The filename into which to write the password generated during setup. 
* `location` : The directory in which to create the motd and password files. If not provided will default to the homedir of the user executing the command
* `shellconfig` : The filename indicating which file to update with the command to dispay the motd on shell instantiation. If not provided, this will default to .bashrc


## Example

```
./clavis -email email@domain.com -name "John Doe" -o ThisCo
```