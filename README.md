# Jemarko Wedding 
this is a wedding website that allows users to rsvp; once they have rsvpd the will be able to select an avatar and leave an optional message, this avatar will join the other guests avatars as they walk around the screen; the avatar plaza will be visible in the backround of the rsvp process. on the 1st of agaust 2026 the rsvp for will close and only the plaza will be visible 


## Requirements:
- the website will be mobile first
- the assets will be hand drawn pngs

### rsvp requirements
- organiser can input the guest list as a csv
- the user will input their name, if in the guest list they will be allowed to input their email
- a confirmation email of who has rsvp'd will be sent to the user
- a user can rsvp on behalf of multiple people

### avatar plaza requirements
- on completion of rsvp the user will be given a selection screen of different avatars, for each user rvpd there should be a selection 
- a user can leave an optionally leave a message, with a 140 character limit 
- useres will be able to see other users avatars with thier messages periodically displaying 
- the plaza will be drawn as the backround to the rsvp process
- users avatars will move around randomly 



## Tech stack 
- Frontend: Svelte (Vercel)
- Backend: Go (Vercel Functions) 
- Database: Supabase 
- Email: Resend 
- Domain: Cloudflare