## MODIFIED Requirements

### Requirement: Character Input Processing
The system SHALL process character input based on the current game mode.

#### Scenario: Classic mode character input
- **GIVEN** the game is in Classic mode and running
- **WHEN** user inputs a letter (a-z, A-Z)
- **THEN** the system SHALL append the lowercase letter to the input buffer
- **AND** check for word matches

#### Scenario: Sentence mode character input
- **GIVEN** the game is in Sentence mode and running
- **WHEN** user inputs a printable character (letter, digit, punctuation, space)
- **THEN** the system SHALL append the character to UserInput
- **AND** compare it with the corresponding target character
- **AND** update correctness statistics

#### Scenario: Ignore invalid input in sentence mode
- **GIVEN** the game is in Sentence mode
- **WHEN** user inputs a non-printable or unsupported character
- **THEN** the system SHALL ignore the input
- **AND** NOT modify UserInput

### Requirement: Backspace Handling
The system SHALL support backspace in both Classic and Sentence modes.

#### Scenario: Backspace in Classic mode
- **GIVEN** the game is in Classic mode with non-empty input buffer
- **WHEN** user presses backspace
- **THEN** the system SHALL remove the last character from the buffer

#### Scenario: Backspace in Sentence mode
- **GIVEN** the game is in Sentence mode with non-empty UserInput
- **WHEN** user presses backspace
- **THEN** the system SHALL remove the last character from UserInput
- **AND** update statistics (decrement total keystrokes)

### Requirement: Enter Key Handling
The system SHALL handle the Enter key differently based on game mode.

#### Scenario: Enter in Classic mode
- **GIVEN** the game is in Classic mode
- **WHEN** user presses Enter
- **THEN** the system SHALL attempt to eliminate a matching word
- **AND** clear the input buffer if successful

#### Scenario: Enter in Sentence mode
- **GIVEN** the game is in Sentence mode
- **WHEN** user presses Enter
- **AND** input length equals target sentence length
- **THEN** the system SHALL finish the game and show results

#### Scenario: Enter before sentence completion
- **GIVEN** the game is in Sentence mode
- **WHEN** user presses Enter
- **AND** input length is less than target sentence length
- **THEN** the system SHALL ignore the Enter key
- **AND** NOT finish the game
