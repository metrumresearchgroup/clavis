# Clavis - The Key

Clavis is a simple binary for auto-provisioning the initial user for Rstudio-Connect. It hinges on the fact that, with password backed authentication, the initial request for creating a user is allowed without an API token. See the below snippet from the Rstudio-Connect Server API Reference:

>This endpoint requires authentication to create other users, which means that you need an API Key for access. How do you get an API Key if there are no users in RStudio Connect?
> * For password authentication, you can use this endpoint without an API Key to create the first user. The first user will be an administrator.


## Parameters

While there is a large array of paramters that can be provided, only three are required to operate:
 * email
 * name
 * organization

 ### Parameter List

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


## Motd
This basically creates a .motd file in the home directory and updates bashrc to cat that file. This is because the built-in MOTD only ever actually gets applied on login, so shells initiated from RStudio or the Desktop terminal emulators will never present them. 

```

  __  __          _
 |  \/  |   ___  | |_  __      __   ___    _ __  __  __
 | |\/| |  / _ \ | __| \ \ /\ / /  / _ \  | '__| \ \/ /
 | |  | | |  __/ | |_   \ V  V /  | (_) | | |     >  <
 |_|  |_|  \___|  \__|   \_/\_/    \___/  |_|    /_/\_\



Welcome to Metworx. RSConnect has been provisioned on this system and your user, darrellb, has been provisioned as an administrator successfully!
Your RSConnect password has been written to /data/home/darrellb/.rsconnectpassword , but you should change it as quickly as soon as you login.

Enjoy Metworx, and enjoy RSConnect!
```