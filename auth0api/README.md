# auth0api

This code is merely a demonstration of how to get and output the existing callbacks, and then allow you to add entries and update them back.

Run this with `updateCallbacks` commented out, capture the output to a `callbacks.yaml`, then edit that file to add the new callback URLs, and comment in the `updateCallbacks`, run `go build`, then re-run `auth0api` to apply the updates.
