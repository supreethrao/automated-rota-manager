# support-bot

Provides a fair rotation algorithm to decide the next person to be on support. It maintains count of the support days of individual team members and uses that to decide based on person having fewest support days. It also takes into account not to put the same person support without a gap of atleast 2 days, irrespective of the number of support days.

## Endpoints

1. GET - `/members` - Lists the details of the current team members in the support rota along with the number of days supported till date and the last date they were on support

2. POST - `/members/:name` - Adds team member into the support rota. The last supported date will be initialised to `31-12-2006` for no real reason other than a date in the past. It won't change any details if the member already exists in the database

3. DELETE - `/members/:name` - Deletes the member from support rota

4. GET - `/support/next` - Evaluates and prints the next person in the rota who should be on support

5. GET - `/support/confirm/:name` - If the person evaluated by `/support/next` is to be confirmed (if not on holiday et al), this endpoint confirms and updates the relevant tables in the database with the details. It's a GET method only to be able to achieve a click and execute functionality. Will print a message saying a person <name> has already been assigned if invoked multiple times on the day.

6. GET - `support/override/:name` - In order to override the set support person for the day (for whatever reason), this endpoint can be invoked and this will change the database details to the new person and adjusts the details of the person who was previously assigned for the day

## TODO

- [ ] Create a namespace to host this application along with having an EBS volume to persist the support information
- [ ] Make it part of a pipeline so that deployment et al can be automated
- [ ] Create a cron functionality so that this app can post in details of the next person on support and links to confirm and override with reasonable details. This will use the slack token and posts into `core-infrastructure` channel everyday at 10 a.m
- [ ] Edit title in the `core-infrastructure` for the person on support details from the `confirm` and `override` endpoints.
- [ ] More endpoints to reset the details of a particular team member
- [ ] Use slack usernames so that the user @ could be used in the slack channel notification


## Known issues
1. When the `override` endpoint is called, the last supported day for the person currently assigned will change to the initial value (`31-12-2006`) which is incorrect. Need to put in the logic to obtain the previous time to today that person was assigned on to support


