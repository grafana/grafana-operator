---
author: "Hubert Stefański"
date: 2023-11-21
title: "Moving Upstream"
linkTitle: "Grafana-Operator Moving Upstream"
description: "An announcement of an upcoming migration to upstream grafana repositories"
---

Exciting times coming up!

The community folks over at upstream Grafana, have reached out and asked if we’d be open to migrating
the Grafana-Operator to their https://github.com/grafana organization! We said "YES!"

## Why?

Ever since we moved away from our initial `Integr8ly` home, (where Peter first created the operator) we have focused on
growing the community and improving the user experience as best we could, we believe that moving to upstream Grafana is
the next logical step.

With this, we hope that the operator gets some more “validity” (for the lack of a better word). With which, we hope new
users and developers will be more likely to engage. Especially given its presence in the official `Grafana`
organization.

It’s not a secret that the current maintainer team isn’t as active as it could be (believe, us, we want to do more).
But... the grafana-operator is a true side-gig for us, 3/4 of our current active maintainers don't work with Kubernetes
day-to-day. We acknowledge that this is one of the major factors limiting development right now. Where some feature
requests have been open for months at a time, with no one being able to implement them. With our limited time we tend to
focus on fixing bugs and small quality-of-life improvements, however, we know that this is not enough and that the
operator has much greater potential. After all, we only support a few core features, out of numerous possible ones in
Grafana.

So, the choice is pretty clear for us, considering the slow development speed, and our limited capacity but also wanting
the operator to grow, in all aspects, community, featureset etc., we believe this to be a step in the right direction
for everyone involved.

## What does it mean for me(as a grafana-operator user)?

Technically, not much will change, the grafana-operator will still be the grafana-operator you’ve gotten to know, love
and have pulled millions of times (we see the stats 😁).

We don’t plan on a new API version with this repo migration, at least not in the upcoming quarters, we’re really happy
with how flexible our API is now, so we don’t really have a reason to change anything.

We also don’t expect existing deployments to be affected in any capacity, and if for any reason that would happen, we'll
be there to help out!

## What about the maintainers?

All current Grafana-Operator maintainers are planning on sticking around and continuing on with our involvement in the
project, after all, this is a massive leap for the operator, and we’re all happy to be part of it!

In all likelihood, new maintainers from the upstream Grafana community might join, so all-in-all, the maintainers might
be more responsive to issues and pull requests!

## Licensing?

We know that news like this often fill developers with dread over commercialization and sudden license changes. Don't
worry - the license will stay exactly the same!

## What WILL change?

Metadata mostly, we’ll likely switch a few key-values in our manifests to better reflect the actual state of the
operator (i.e ownership, repository addresses, maintenance contacts etc)
In general, house-keeping stuff that doesn't really affect users.

The primary change will be the repo address on Github, we'll migrate
the https://github.com/grafana/grafana-operator to https://github.com/grafana.
Github automatically redirects migrated addresses, so there's nothing to be concerned about on that end.
Future OCI URI's will reflect this change, while existing artifacts will still have the same URI, they will be copied
over to our new home with the caveat, that digests will change.

The documentation URL will also change, however, we will set up a redirect from it's current location, so it's likely
you won't even notice it moved!

Perhaps, the biggest change to you, as a contributor to the grafana-operator, would be a new requirement to sign the
[Grafana CLA](https://grafana.com/docs/grafana/latest/developers/cla/), But worry not, this CLA is based on the
Apache Software Foundation CLA, which should put those of us whom are concerned about signing agreements, at ease.

We might end up moving a few other things around, like creating a new channel on the upstream Grafana slack, however,
you'll definitely hear about any other changes before they happen! So, stay tuned!

## Summary

Thanks to everyone that supported us over the past 4 years, the project has really grown beyond expectation. From a
niche operator created for a very specific project, to a pretty sizeable community and a project that is now used by
many major companies throughout the world, some of whom you wouldn't expect, as they don't announce it publicly, but
contact us for support!

We've got our hopes set high for this migration, and we hope you do too, let's take the operator to the next level!

Thanks ~ Grafana Operator Maintainers
