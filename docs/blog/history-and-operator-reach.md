---
author: "Hubert Stefański"
date: 2023-12-05
title: "Grafana-Operator - A small subproject that made it big"
linkTitle: "Grafana-Operator - A small subproject that made it big"
description: "A History of the grafana operator and its growth, and why it's so 'quiet' "
---

# Grafana-Operator - A small subproject that made it big

This blog post will describe the journey that the grafana-operator underwent, from a small subproject from within a Red
Hat product offering, to one silently being used by some of the biggest companies worldwide, as well as where it's going
to next!

## The Origin

**Disclaimer: Most of the information written in this section is from before my time being involved with the project,
written to the best of my knowledge.**

The Grafana-Operator was initially created as part of the monitoring stack used in Red Hat Managed Integration (RHMI),
or "Integr8ly" for its open-source name. Peter Braun created it back in 2019, giving it its start in open source.

Unsurprisingly, the feature development for the operator was driven mostly by the requirements derived from Integr8ly,
that is, simple management of dashboards and datasources (ironically saying "simple" is oversimplifying the work that
went into it). And as time would show, these bare few requirements were something that a multitude of other teams and
companies also had. This serves as a prime example of how a relatively small overhead in operator development can yield
massive benefits for its users, granted, we didn't know how big the eventual user base would be, at the time.

## Growth - How and Why?

Slow and steady, is the most suitable description of the growth of the operator over the past three years. Granted, this
assertion comes mostly from what we see in terms of stars/visitors and contributions to our git repository, that being
pretty linear over time. However, judging the true size of an open source project based on these criteria isn't entirely
the most accurate. Let me explain how and why that is, and why the project is in fact much bigger than it would appear
by just browsing our git repository.

### How did the operator grow?

I don't think there's a special recipe in how we grew the operator, both in terms of features and community.
I guess as long as you just make stuff that works, try not to break anything from version to version (Easier said than
done, right?), and respond to user questions and support, then that's all that it takes?
Our development story really isn't that complicated, most of it was prompted by user questions, feedback (complaints as
well) and generally that's what we did.

One of the key milestones was definitely V5, where we decided to completely re-write the operator using a new approach.
Which we've done in hopes of improving the developer experience, as previous versions suffered greatly from code creep,
pretty much every single controller was written in a different style, with different ways of handling the same cases,
but for different resources. And that was something which really blocked potential contributors from, well,
contributing. Even the core maintainer group would have to constantly refresh their memory on how and why something
worked the way it did.
Realising this, was definitely a first step in the right direction, at least setting the foundations for the eventual
"upstreamification".

## Why is the Grafana Operator so widely used, yet relatively small on GitHub?

As I've said in the introduction of this blogpost, the operator, in my opinion, is a "quiet giant". That could be due to
a number of reasons:
The following list is just observations, not complaints ;)

### 1. The git repository itself is largely meaningless when it comes to a vast majority of the users.

This is not to say that grafana-operator users don't care about the operator, but rather, they don't need to care, and
ironically, I see this as a pretty positive sign that the project is in good shape. This is mainly because most of the
users opt to install the operator through <insert the most popular installation method in whatever year you're reading
this>. Joking aside, users value ease of use, so it's no indictment against anyone for just wanting stuff to work, be it
they install it through OLM/Operator Hub/Helm/whatever else. Frankly, this is the way the vast majority of users will
interact with the operator, rather than installing from source through our git repository, there simply is no reason for
them to do so.

The convenience of OLM/OperatorHub and Helm means that that's where a potential user will first come across an operator.
A fitting comparison would be that you'll probably go to buy milk in the nearest shop, rather than find the farm from
which it originates, does that mean you like your milk more or less? You probably don't care!

For one, this makes it easier than ever to just get an operator out there, and have it gather users.

### 2. Variety of support channels

Generally, we try to point people to existing issues (which we try to gradually work through and close, time permitting)
when they inevitably arise. However, synchronous human interaction seems to be the way most people prefer to resolve
their issues nowadays, and that's also valid. Most of our issues tend to be reported through our k8s.io Slack based
channel, and close to 90% of those just tend to be general configuration and PEBKAC errors (yes, we can always do better
on the docs side! So really, it's our fault in the end).

The tendency to use Slack as the primary support channel definitely means less engagement on the repository itself,
however, yet again we can't blame a user for doing what's easiest for them! But it does mean that we have roughly twice
as many Slack channel members, than we do stars on the project.

There's another aspect to this, that is largely a virtue of how RH does business, Peter and I, often get contacted
through our corporate email from Red Hat Technical Account Managers and consultants, with regard to support queries from
their respective customers, and we also answer a great deal of cases on that end. The customers of whom are often high
profile. I'll expand on this later on in this blog post.

### 3. If it works, you just don't hear about it as often (as a developer)

Generally, people don't come to open source repositories to simply praise the project, although, there sometimes are
those kind few souls that do! By that nature, it's more likely that we'll have a user find us to ask for support/report
a bug or something else of that nature, rather than to simply leave a star on GitHub.

### 4. We don't really advertise the operator as much as some others might

The good old adage about a great project being only as good as how you can sell it holds true. The core maintainer group
isn't really that social, We've done a few presentations here and there, written a few blogposts, but it's unlikely
you'll find any of us, posting about the operator every day on LinkedIn etc. (Edvin generally posts on major
milestones!)
This is both likely a characteristic of our nature as software engineers, and the fact that this really is a side-gig
for most of the maintainer group.

Most of the maintainers on the project aren't really involved with Kubernetes/OpenShift on a daily basis anymore. For
example, it's been close to 3 years now, that working with OpenShift was one of my main responsibilities, while now,
outside the Grafana-Operator, this type of work shows up once every few weeks, at best.

### So, why is the Grafana-Operator a "quiet giant"?

This is mainly down to the fact that rarely any Grafana-Operator users will visit our git repository, as most will
install through OLM/OperatorHub, so in reality our best guess is down to how many container image downloads happen on a
daily/weekly/whatever basis (if you know how we could get these metrics from OLM, or others, please let us know!).

From our available metrics, V5 has been downloaded just over **3.1 million** (as of November 27th) times since **Jun 9th
2023** (the first available V5 branch release) giving an average of **18k daily downloads**! Which isn't an
insignificant
number!

As mentioned previously, we get quite a few private emails/messages, from users whom don't exactly want to advertise
that they're using a specific project. Some of who simply don't advertise regardless.
Within this "group" of users, we've got to know sizeable banks (National and International ones), Automotive
manufacturers, Stock Exchanges, cloud providers etc., the list goes on and on.
I won't divulge the specifics of these users, but our list of known, or assumed users (based on who forked and starred
our repository) is pretty large and varied.

Granted, a lot of the success is due to the popularity of Grafana itself, but the added benefit of the "
operationalization" of the management of Grafana instances is something that has significant value for users. We're well
aware of the fact that many of our users have enterprise contracts with Grafana, and we've also adapted the operator to
be able to manage resources on "external" Grafanas. Which now means, you can still have an enterprise contract with
Grafana, but allow your monitoring team to have a GitOps-based approach to defining your monitoring stack, without
having to manage the Grafana instance yourself (be it through the operator, or through a managed Grafana as a service).

If I were to express my personal thought on why users find the operator valuable, it would probably be exactly because
the operator just makes it easier to manage Grafana at scale, in a workflow that is familiar to many DevOps/platform
engineers.
As with all things, the management of software is a balance of compromises, in order to reap the benefit of a piece of
software, you have to accept some of the overhead that comes with it. And at the end of the day a user is most likely to
go with a solution that works well for their use case, and doesn't introduce a complex layer unnecessarily. I firmly
believe that the Grafana-Operator classifies itself as one of these projects, we add a small overhead (you have to use
our Custom Resource Definitions as a wrapper for your Grafana resources), but we make up for it in the overhead we
remove from the operations side.

The paragraph above massively simplifies the entire debate and decision process that goes into selecting a bit of
software, as it rarely is as clear-cut as "just choosing the easiest solution". However, I would still stand by the
sentiment, that the balance of compromises falls in favour of using the Grafana-Operator over self-managed instances.

## Moving upstream

It really isn't a secret that the operator is in a bit of a stagnation right now, which the maintainers are well aware
of. Our day-to-day jobs take a significant amount of time and energy, and most of what we are able to dedicate to the
operator is a 30-minute call once a week, and answering questions/issues on our GitHub on a best effort basis. However,
it is important to note that the move upstream is not a result of us wanting to offload the maintenance of the operator
to someone else. All maintainers currently active in the project are planning on staying around and continuing our
involvement.

Internally, within our small maintainer group, we had a few conversations around what could be the next major step for
the operator. We did bounce the idea of reaching out and maybe having the project be adopted by the CNCF or Grafana.
However, we never really made it a hard goal or made any steps towards either one of those options.
Luckily, upstream Grafana folks reached out to us first, (which was a very welcome surprise to us). Seeing the
initiative from upstream in adopting the operator and migrating it into the official Grafana Labs GitHub organization
was a great motivational booster. For numerous reasons (which I'll clarify below) this seemed like the most logical step
for the operator, so we fully embraced the idea.

### Why move?

I know, this question is self-explanatory, almost all open source projects would look forward to being somehow
incorporated or acknowledged to a somewhat "official" extent upstream.
Although, I think it's important to highlight the reasons why it is a logical choice.

As a general rule I believe all the points I'm about to make below are based on wanting the best for the operator and
its users. We believe that for the operator to continue to grow, be successful and continue to deliver value to users,
it must improve in community engagement (even though we have gone a long way, there's still a ways to go!)

Being a downstream project does have its downsides, as a software engineer I often share the same sentiments that I know
others have. That is, there's always an element of "carefulness" when proposing or integrating a new project which is
relatively unknown or hosted in a small repository. We know that feeling, it's not based on "distrust" per se, but it's
just being hesitant about a project that doesn't really have "apparent validity'.
Moving into an official Grafana GitHub organization would give the operator this "validity". Deep down, all this really
means is that users might be more convinced to use it based on the sheer fact that it's hosted in a known, public and
popular repository.

An extension to the point above, is that, it's hard to grow a community, when the community doesn't really know how to
find you, meaning that there's a slope of sorts, which without "backing" from an official organization is hard to
overcome.

Community growth is what keeps the operator alive, and we don't see a better way of facilitating that growth without
associating ourselves with an upstream. This comes with a range of possible benefits (whether those will come to
fruition, time will tell). By becoming part of the Grafana organization we can start to offer more to existing users, by
now being more involved and closer to the actual product on which we operate.
As the community grows, so will the feature set, as demand and the number of possible contributors creating these
features increases. This will also bring clearer goals into the operators' development path, the community will make its
expectations known, and the eventual development can be led by those expectations, rather than anticipation of user
needs. Which is mainly how development has been happening up until this point. "Mainly", but not "solely".

The move to upstream is prompted by both acknowledgement of accomplishment of what the operator has done up until this
point, after all it must have a real benefit if Grafana wants to "adopt" it. As well as the acknowledgement of current
stagnation, where the current maintainers cannot devote time and effort consistently to meet user needs, despite best
efforts and intentions.

All-in-all, the decision is driven purely by a desire to grow and improve, for the sake of the community and our users.

## Personal Reflection

My engagement with the operator began by trying to fix a bug in for my previous team and project, and by pure chance I
happened to spot a few areas that could be improved, so I began contributing those fixes. Back then, I couldn't have
imagined where this would land the operator, and how I'd at least put a few bricks to that.
The work albeit sometimes feeling unimportant, or insignificant in terms of the size of contribution, definitely has its
rewarding side, when a company/user reaches out and says "hey, thanks, we find it useful", or "thanks for fixing that".
I've had the chance to interact with a wide variety of people, from within and outside of Red Hat, everyday DevOps
engineers, tech leads, architects, consultants, TAMs , and CTOs of sizeable companies, all of whom at some point needed
help with some aspect of the operator.

I look forward to continuing this journey, and I am super thankful for all the other maintainers that make it all happen
day-to-day:
**Peter**, **Edvin** and **Igor**, and hopefully many more to come!

Extra reading: https://grafana-operator.github.io/grafana-operator/blog/2023/11/21/moving-upstream/ <- For more context

Feel free to reach out if you've got any questions or feedback!
