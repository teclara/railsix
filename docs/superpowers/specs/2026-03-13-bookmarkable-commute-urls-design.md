# Bookmarkable Commute URLs

## Problem

The commute dashboard at `/` stores route configuration in localStorage only. Users cannot bookmark specific routes, share links, or save separate morning/evening bookmarks.

## URL Scheme

```
/?from={stopCode}&to={stopCode}&dir={toWork|toHome}
```

All three params required for URL mode to activate. Missing or invalid params fall back to localStorage behavior.

**Examples:**
- Morning commute: `/?from=UN&to=OASH&dir=toWork`
- Evening commute: `/?from=OASH&to=UN&dir=toHome`

## Behavior

### URL Params Present

- URL params drive the view — origin, destination, and direction toggle state
- localStorage is NOT read or modified — URL is a temporary view layer
- Toggle button swaps `from`/`to`, flips `dir`, and updates the URL via `goto()` with `replaceState`
- Page title updates: "UN → OASH — Rail Six"

### Bare `/` (No Params)

- Current behavior unchanged: localStorage + time-of-day auto-switch
- Page title: "Rail Six"

### Toggle Behavior (URL Mode)

When user clicks the inactive direction toggle:
1. Swap `from` and `to` values
2. Flip `dir` (`toWork` ↔ `toHome`)
3. Update URL via `replaceState` (no page reload)
4. Fetch departures for the new origin/destination pair

### Validation

- `from` and `to` must match known stop codes from the stops list (loaded server-side)
- If either code is invalid or any param is missing → ignore all URL params, fall back to localStorage
- `dir` must be `toWork` or `toHome` — invalid value → fall back

## Implementation

### Files to Modify

1. **`src/routes/+page.server.ts`**
   - Read `from`, `to`, `dir` from `url.searchParams`
   - Validate stop codes against the stops list
   - Pass validated `urlTrip: { from, to, dir } | null` in page data

2. **`src/lib/components/MyCommute.svelte`**
   - Accept optional `urlTrip` prop
   - If `urlTrip` is present: use it for origin/destination/direction instead of localStorage
   - Toggle updates URL via `goto()` with `replaceState: true`
   - All departure fetching and display logic works the same — it just reads from URL state instead of commute store

3. **`src/routes/+page.svelte`**
   - Pass `urlTrip` from page data to `MyCommute`
   - Update `<title>` tag when `urlTrip` is present

### Files NOT Modified

- `src/lib/stores/commute.ts` — URL params bypass the store entirely
- `src/lib/api.ts` / `src/lib/api-client.ts` — no API changes needed
- Backend services — no changes

## Edge Cases

- User has localStorage commute AND opens a URL → URL wins for that view, localStorage preserved
- User opens URL then navigates to bare `/` → back to their localStorage commute
- User with no localStorage opens a URL → works fine, shows the URL route
- User with no localStorage visits bare `/` → shows CommuteSetup as before
