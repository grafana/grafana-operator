# Security Policy

If you believe you've found a security vulnerability in this operator then please follow this document, we do our best
to address issues ASAP.

## Supported Versions

The use of 2.x.x is strongly discouraged, it is not supported or updated, versions prior to 3.x.x have no clear upgrade 
path to the current release, any security vulnerabilities in 2.x.x releases won't be addressed, unless done so by a
member of the community who uses that specific version,
please use the latest release of this operator to ensure most up-to-date code, this also means that we can address and 
find vulnerabilities quicker.

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :x:                |
| 3.x.x   | :white_check_mark: |


## Dealing with and reporting a Vulnerability

Ideally we want to resolve vulnerability issues as soon as possible, to help us with this please consider doing the 
following:

If a vulnerability has been found, please investigate further so that you can provide as much context as possible when 
reporting, specifying affected components/libraries/versions etc. this will help a great deal in resolution.

During your investigation if you happen upon a fix, please create a pull request with the proposed solution to that
vulnerability, detail the problem, your findings and the solution provided. By doing this, we can address the issue
sooner, having only to review and verify that the solution works as opposed to having to investigate this separately and 
take more time to reach a solution.

If there is no clear fix that was found during your investigation then please notify the repository admin(s) with the
 findings of your investigation directly, ideally by email, we want to avoid making security vulnerabilities public 
 before we can solve them, preventing bad actors from exploiting users of this operator.
 
If the security vulnerability exists in Grafana itself then please bring that up with the upstream
Grafana repository adhering to their doc [here](https://github.com/grafana/grafana/blob/master/SECURITY.md) and notify
the admin(s) of this operator to update when a patch has been released with the specific vulnerability fix.


## Contact
Please contact the following people with information regarding found vulnerabilities:

@pb82 - pbraun@redhat.com

@david-martin - davmarti@redhat.com