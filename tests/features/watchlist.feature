Feature: Watchlist

  Scenario: Add a stock to watchlist
    Given a user has no watchlist entries
    When they add AAPL to the watchlist
    Then GET /api/watchlist returns AAPL

  Scenario: Add duplicate watchlist entry
    Given a user already watches AAPL
    When they add AAPL again
    Then the request fails with 409
