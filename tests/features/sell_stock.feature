Feature: Sell stock

  Scenario: Sell all shares
    Given a user holds 5 AAPL
    When the user sells 5 AAPL
    Then the holdings row is deleted and a SOLD transaction is recorded

  Scenario: Sell more shares than owned
    Given a user holds 1 AAPL
    When the user tries to sell 10 AAPL
    Then the request fails with 400 and their holdings are unchanged
