# Task

YOU are creating an expense tracking app which will work on a google pixel 9, so you have access to a TPU and basic on device AI. There is already a plan in BACKEND_API.md and FRONTEND_API.md... the backend will be built in golang... There will be a android app which will read notifications/messages, and do a basic classifications of all bank transactions. there will also be a web app which will help the user see the graphs etc easier, but it will work on the android device as well..


## Core Functionality
Similar to how truecaller works with its notifcations -- reads the notifications, get the merchant. if classification isnt possible, then a basic location should be recorded... the android app will do a lot of the heavy lifiting tbh. the backend will just be to store. and during the day, it will all be on the local storage of the app, and will be synced using rsync to google drive.


## Development process

- When making any changes for any task, in ./progress/{task}.md, keep an updated log of your todo list, your plan for executing the todo list as a basic overview for a reviewer. keep checking off/adding/updating items as you come across more things you need to account for while working on the task at hand. Only start working on a task after the initial plan stored in progress has been reviewed and accepted. This rule is non-negotiable for anything where you would be making code changes.
