Feature: Buy stock

  Scenario: Buy stock with enough balance
    Given a user with balance $1000 and live price of AAPL is $200
    When the user buys 2 AAPL
    Then their balance is $600, they hold 2 AAPL, and a BOUGHT transaction is recorded

  Scenario: Buy stock with insufficient balance
    Given a user with balance $100 and live price of AAPL is $200
    When the user buys 1 AAPL
    Then the request fails with 400 and their balance is unchanged

  Scenario: Buy stock increments an existing holding
    Given a user already holds 3 AAPL
    When the user buys 2 more AAPL
    Then they hold 5 AAPL
