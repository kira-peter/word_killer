## ADDED Requirements

### Requirement: Game Mode Enumeration
The system SHALL define a `GameMode` type to represent different game modes.

#### Scenario: Mode type definition
- **GIVEN** the game package is imported
- **WHEN** accessing the GameMode type
- **THEN** it SHALL provide constants for Classic and Sentence modes
- **AND** the type SHALL be an integer-based enumeration

### Requirement: Mode Selection and Initialization
The system SHALL support selecting and initializing different game modes.

#### Scenario: Classic mode initialization
- **GIVEN** a game instance
- **WHEN** `Start()` is called with word count parameter
- **THEN** the game SHALL initialize in Classic mode
- **AND** load words from the configured dictionaries

#### Scenario: Sentence mode initialization
- **GIVEN** a game instance
- **WHEN** `StartSentenceMode()` is called
- **THEN** the game SHALL initialize in Sentence mode
- **AND** randomly select one sentence from the sentences file
- **AND** store the sentence as the target

### Requirement: Sentence Data Loading
The system SHALL load sentences from a text file for Sentence mode.

#### Scenario: Load sentences successfully
- **GIVEN** a valid sentences file path
- **WHEN** `LoadSentences(path)` is called
- **THEN** the system SHALL read the file line by line
- **AND** return a slice of non-empty sentence strings
- **AND** each sentence SHALL be trimmed of leading/trailing whitespace

#### Scenario: Handle empty or missing file
- **GIVEN** an invalid or empty sentences file path
- **WHEN** `LoadSentences(path)` is called
- **THEN** the system SHALL return an error
- **AND** provide a descriptive error message

### Requirement: Sentence Mode Input Handling
The system SHALL handle character input differently in Sentence mode.

#### Scenario: Accept valid sentence characters
- **GIVEN** the game is in Sentence mode and running
- **WHEN** user inputs a letter, digit, punctuation, or space
- **THEN** the system SHALL append the character to UserInput
- **AND** increment keystroke statistics

#### Scenario: Reject invalid characters
- **GIVEN** the game is in Sentence mode and running
- **WHEN** user inputs a control character or unsupported symbol
- **THEN** the system SHALL ignore the input
- **AND** NOT modify UserInput

### Requirement: Sentence Mode Completion
The system SHALL determine game completion based on sentence length in Sentence mode.

#### Scenario: Complete sentence typing
- **GIVEN** the game is in Sentence mode
- **WHEN** user input length equals target sentence length
- **AND** user presses Enter
- **THEN** the game SHALL transition to Finished status
- **AND** calculate final statistics (total characters, correct characters, accuracy)

#### Scenario: Allow errors in completion
- **GIVEN** the game is in Sentence mode
- **WHEN** user input contains errors but reaches target length
- **AND** user presses Enter
- **THEN** the game SHALL still complete
- **AND** reflect errors in accuracy statistics

### Requirement: Sentence Mode Character Matching
The system SHALL track character-by-character correctness in Sentence mode.

#### Scenario: Match correct character
- **GIVEN** the game is in Sentence mode
- **WHEN** user inputs a character that matches the corresponding target character
- **THEN** the system SHALL increment correct character count
- **AND** mark the position as correct for UI rendering

#### Scenario: Detect incorrect character
- **GIVEN** the game is in Sentence mode
- **WHEN** user inputs a character that does NOT match the corresponding target character
- **THEN** the system SHALL NOT increment correct character count
- **AND** mark the position as incorrect for UI rendering
