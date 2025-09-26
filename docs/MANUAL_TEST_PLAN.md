# VidFriends Manual Test Plan

This guide outlines the manual end-to-end scenarios required to validate the VidFriends MVP experience. Each flow should be executed on the current staging or local development environment using the latest main branch build.

## Test Environment

- **Backend**: Go API running against a PostgreSQL instance with migrations applied.
- **Frontend**: React application served via `npm run dev` or the staging deployment.
- **Seed Data**: Ensure the default seed users and sample videos defined in the migrations are loaded. Create additional accounts as needed during testing.
- **Test Accounts**: Use dedicated testing emails (e.g., `test+<scenario>@vidfriends.app`) to avoid conflicts with production data.

## Signup Flow

1. Navigate to the signup page.
2. Enter a unique email address and valid password meeting the documented requirements.
3. Submit the form.
4. Follow any verification steps (email link or OTP) if configured.
5. Confirm the user is redirected to the onboarding/home experience.

**Expected results**
- New user record is created in the database and visible through the admin/inspection tooling.
- Session cookie or token is issued and stored in the browser.
- Onboarding checklist or feed loads without errors.

## Login Flow

1. From a logged-out state, open the login page.
2. Enter credentials for an existing user account.
3. Submit the form.
4. Wait for the dashboard/home feed to load.

**Expected results**
- API responds with a success status.
- Refresh token is stored and access token is usable for subsequent API calls (verify via developer tools network tab).
- User avatar, friends list, and feed render without placeholder data.
- Logout control becomes available.

## Friend Invitation Flow

1. Ensure two accounts exist: **Inviter** and **Invitee** (both verified and able to log in).
2. Log in as the Inviter.
3. Open the friends or invite management screen.
4. Send an invite to the Invitee using their email or username.
5. Confirm a success toast or notification appears.
6. Log out and log in as the Invitee (or open a separate session/browser profile).
7. Navigate to the pending invites list.
8. Accept the invitation.

**Expected results**
- Invite is persisted server-side and visible in the database.
- The Inviter sees the invite status update to “Accepted” without manual refresh (optimistic UI succeeds or rolls back appropriately).
- Both accounts show each other in the friends list after acceptance.
- Relevant activity feed entries or notifications display.

## Video Sharing Flow

1. Ensure Inviter and Invitee accounts are friends.
2. Log in as the Inviter.
3. Open the video share interface.
4. Provide a valid video URL that `yt-dlp` can process.
5. Add optional notes/tags and select the Invitee as a recipient.
6. Submit the share request.
7. Verify the video appears in the Inviter’s outgoing shares list with a processing state.
8. Wait for background processing to complete and confirm the status transitions to available.
9. Log in as the Invitee.
10. Check the incoming shares list for the new video.
11. Play the video or open the metadata modal.

**Expected results**
- Share record is created and linked to the selected friends.
- Background job uploads assets to object storage and marks the share ready.
- Invitee receives a notification or badge for the new share.
- Video metadata and playback load successfully without console errors.

## Regression Checks After Each Flow

- Refresh the page to ensure session persistence via refresh tokens works.
- Inspect logs for warnings or errors generated during the flow.
- Attempt invalid inputs (duplicate invites, unsupported URLs, weak passwords) to confirm validation messages display.

## Reporting

Log results in the shared QA tracker or ticketing system with the following details:

- Tester name and date
- Environment (staging/local + commit hash)
- Scenario executed
- Outcome (pass/fail)
- Screenshots or console logs for failures
- Follow-up tickets filed (if applicable)

Execute this plan before major releases or when significant backend/frontend changes land to maintain confidence in the core user journeys.
