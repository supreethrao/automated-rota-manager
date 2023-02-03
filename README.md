# automated-rota-manager

Provides a fair rotation algorithm to decide the next person in the rota. 
It maintains count of the number of days an individual team member has been picked and uses that to decide the next person with least number of days. 
It also takes into account not to pick the same person without a gap of at least 2 other picks, irrespective of the number of days the person has accrued.

## Project setup
Makefile has a target `setup` which should setup the project. 
This project uses ginkgo/gomega for testing and `make check` should run the testsuite. 


## Endpoints

1. GET - `/members` - Lists the details of the current team members in the rota along with the number of days accrued till date and the last date they were picked.

2. POST - `/members/:name` - Adds team member into the rota. The last picked date will be initialised to `31-12-2006` for no real reason other than a date in the past. It won't change any details if the member already exists in the database

3. DELETE - `/members/:name` - Deletes the member from rota

4. GET - `/rota/next` - Evaluates and prints the next person in the rota

5. GET - `/rota/confirm/:name/:date` - If the person evaluated by `/rota/next` is to be confirmed (if not on holiday et al), this endpoint confirms and updates the relevant tables in the database with the details. It's a GET method only to be able to achieve a click and execute functionality. Will print a message saying a person <name> has already been assigned if invoked multiple times on the day.

6. GET - `/rota/override/:name` - In order to override the person picked for the day (for whatever reason), this endpoint can be invoked and this will change the database details to the new person and adjusts the details of the person who was previously assigned for the day. Override is always for the current day.

7. POST - `/outofoffice/:name/:from/:to` - Records the out of office dates for a person. The from and to should be in the format `DD-MM-YYYY`. The person out of office will be skipped from rota. The to date is one day before the return date.

8. GET - `/outofoffice` - Gets the out of office schedule for the team

9. GET - `/outofoffice/:name` - Gets the out of office schedule for the specific team member
