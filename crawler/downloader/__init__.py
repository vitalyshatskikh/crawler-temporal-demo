"""
Downloading is a complex task, including session management (headers, cookies), proxy usage,
downloading the main content, and/or following additional links.
Retry attempts may also be necessary due to limitations on the source site.

Therefore, the downloader is implemented as a separate Temporal-based application
"""