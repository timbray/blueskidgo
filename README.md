# Blueskidgo

@bluesky Identity tooling in Go

Miscellaneous infrastructure for the **@bluesky Identity** 
scheme as proposed [here](https://www.tbray.org/ongoing/When/202x/2020/12/01/Bluesky-Identity).

Currently the short-term goal is to sketch in enough of the
protocol to illustrate it for the purposes of the 
[Satellite bluesky contest](https://blueskyweb.org/satellite).

### The Server

At the top level, `blueskid.go` contains a small http server
that listens on port 8123 by default, but you can change
that with the `--port` option. Let's just call this the 
Server.

### Identities

Accounts on Providers such as Twitter and Reddit are 
called *Provider Identities* (PIDs),
and use syntax such as `twitter.com@timbray` and
`reddit.com@timbray`.

A *Bluesky Identifier* (BID) is a higher-level construct 
which may be used to map together multiple PIDs so they 
can be considered a single source, for purposes such as 
reputation metrics. In this implementation, a BID is 
represented by a 64-bit unsigned integer, or alternately
by a 16-character hex string (in all-caps) giving its value.

### Assertions 

The Blueskid protocol relies on embedding assertions in
the text of social-media posts.  These assertions have a small number 
of string fields.

The beginning and end of an assertion are marked by "🥁" 
(U+1F941 DRUM) and the fields are separated by "🎸"
(U+1F3B8 GUITAR).  It would probably be better to use 
two different charactures to mark the start and end
of the assertion. Regrettably there is no harmonica emoji.

There is a possibility that a PID might include 🥁 or 🎸, 
so in all assertions that contain a PID, that PID 
appears in the last field, to allow the use of libraries such as 
Go's `strings.SplitN`, which specify the maximum number 
of fields.

The first field of every assertion is a single character 
identifying the type of assertion. "C" means Claim BID, 
"G" means Grant BID, "A" means Accept BID, and "U" means 
Unclaim BID.

### Claiming a BID

The Server can generate a BID Claim assertion. To do this,
send a `POST` to the `/claim-assertion` endpoint as follows:

```json
{
  "BID": "309F0000021"
}
```
The BID should be provided in hex.

Assuming nothing goes wrong, you'll get back a JSON
construct that looks something like this:

```json
{
  "ClaimAssertion": "🥁C🎸309F0000021🥁"
}
```
### Sharing BIDs between PIDs

The Server can generate a pair of assertions by which a PID 
on a @bluesky Provider can grant shared ownership of a *Bluesky 
Identity* (BID) to another account on the same 
or another Provider.  To do this, send a `POST` to 
the `/grant-assertions` endpoint as follows:

```json
{
  "BID": "309F0000021",
  "Granter": "twitter.com@tim",
  "Accepter": "reddit.com@tim"
}
```

The BID should be provided in hex.

Assuming nothing goes wrong, you'll get back a JSON
construct that looks something like this:

```json
{
  "GrantAssertion": "🥁G🎸309F0000021🎸eF2QINuVp9Q=🎸MCowBQYDK2VwAyEAj9Z3Lf5Rxylw6WParFBmeSnyhb7rK4+n1QsQba1OX2Q=🎸a/N23VuG3n7p0lfUbfPxzdDb0Ur81S3vThG0x1ZoLtf8eUHP+4AD6sOVEkx2nPkmGyMUfTyPzUcTZ/HvGs08CA==🎸reddit.com@tim🥁",
  "AcceptAssertion": "🥁A🎸309F0000021🎸CJHlLHY9das=🎸MCowBQYDK2VwAyEAj9Z3Lf5Rxylw6WParFBmeSnyhb7rK4+n1QsQba1OX2Q=🎸rnwypUgFm5YmmFVxsh8mTInvAeAUxET8lUVId9OU9cR9wtfMWyXVDMkQyVnHoCqnUSn18+9HGr2gEF7lXOwYDg==🎸twitter.com@tim🥁"
}
```

### Unclaiming a BID

The Server can generate a BID Unclaim assertion. To do this,
send a `POST` to the `/unclaim-assertion` endpoint as follows:

```json
{
  "BID": "309F0000021"
}
```
The BID should be provided in hex.

Assuming nothing goes wrong, you'll get back a JSON
construct that looks something like this:

```json
{
  "UnclaimAssertion": "🥁U🎸309F000021🥁"
}
```

### Verifying assertions

To process a grant of a BID from PID to PID, it is necesary
to validate a pair of assertions - one from the granter, one
from the accepter - very carefully. 

This includes verifying the signatures using the provided
public key to prove that the creator of the posts containing
the Grant and Accept assertions was at one point in time 
in possession of the private key that was used to generate
both. 

There are several other sanity checks in the function
`checkGrantAssertion` and since I'm not a crypto weenie, I 
probably missed a few that need to be added.

### Retrieving assertions 

@bluesky Identity assumes that assertions claiming and 
sharing BIDs will be posted to social-media Providers, 
for example Twitter.  Retrieving data from Providers
is sufficiently idiosyncratic that custom code is required
for each.

Code in `twitter.go` uses the V2 Twitter API to retrieve a tweet
containing a blueskid assertion and unpack it. 

These days, you can't just do an HTTP GET on a tweet URL
and receive the content. So to use this, you need to get 
a Twitter Developer Account 
approved, retrieve a bearer token, and arrange for the
`TWITTER_BEARER_TOKEN` environment variable to have that 
value.

`tumblr.go` and
`mastodon.go` make a best-effort to pull the assertion out
of the jumble of HTML this kind of site produces.

### Cryptography

This software uses only ed25119 (EdDSA) keys.

`ed25519.go` provides utilities for converting public keys
back and forth between string and binary representations.
This uses the horrible old ASN.1/PEM/PKIX machinery, which
would be silly if the whole world used Go, but many other
popular libraries in popular languages assume this is the 
one and only way to interchange public keys. Thus this is 
the right
thing to do in an Internet Protocol.  At least you don't 
have to think about it.

### The Ledger

The @bluesky Identity protocol requires the presence of
a Ledger, to which a record of Claim, Unclaim, and Grant
BID transactions are committed immutably.  

The Server implements (see `ledger.go`) an ephemeral ledger 
that is a fake, lives only in memory and is not persisted. 
Databases are hard and this is just a demo!

However, the API offered by the Server for updating and 
scanning the ledger constitutes a proposal for what the
API for a less-fake ledger must look like.

When a BID Claim assertion has been posted, send a POST to
the `/claim-bid` endpoint as follows:

```json
{
  "Post": "url of social-media post containing the BID claim assertion"
}
```
There is no response body.

When a BID Grant assertion and corresponding BID CLaim 
assertion have both been posted, send a post to the 
`/grant-bid` endpoint as follows:

```json
{
  "GrantPost":  "url of social-media post containing the BID-claim assertion"
  "AcceptPost": "url of social-media post containing the BID-accept assertion"
}
```

There is no response body.

When a BID Unclaim assertion has been posted, send a POST to
the `/unclaim-bid` endpoint as follows:

```jaon
{
  "Post": "url of social-media post containing the BID unclaim assertion"
}
```
There is no response body.

### Ledger records

Each ledger record has four fields. 

"RecordType" must be 
one of "Claim", "Grant", or "Unclaim". [Actually, in the 
current implementation the values are 0, 1, and 2.]

"BID" must be a hex encoding of the 64-bit BID being
transacted. 

"PIDs" is an array with one or two members. In
Claim and Unclaim records, it has one element giving the
PID claiming or unclaiming.  In a Grant record it has
two elements giving the granting and accepting PIDs.

"Posts" is an array with one or two members. In Claim and
Unclaim records, it has one element giving the 
URL of the social-media post
containing the assertion.  In a Grant record it the 
first element is the URL of the social-media post 
containing the Grant assertion, the second the URL of 
the social-media post containing the Accept assertino.

To get a JSON dump of the current status of the ledger, 
do a GET on the `/ledger` endpoint.

### The database

When the ledger is updated, the server updates internal
tables containing the mappings between BIDs and PIDs.  This 
is fairly necessary, because to claim a PID, there needs
to be a check that it wasn't claimed by someone else.

There are three endpoints provided to query the database. 
`/bids-for-pid` takes a single query parameter named `pid` 
and yields a JSON list of the BIDs mapped to that PID.

THe inverse service is provided by `/pids-for-bid`, which
takes a single query parameter named `bid`.

Finally, the `pid-group` endpoint, which takes a single
query parameter `pid`, yields a list containing this PID 
and all other PIDs that are mapped to it through one BID or
another.

