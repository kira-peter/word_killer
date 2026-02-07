## ADDED Requirements

### Requirement: Mode Selection Screen
The system SHALL provide a mode selection screen after the welcome screen.

#### Scenario: Display mode options
- **GIVEN** the user selects "start" from the welcome screen
- **WHEN** the mode selection screen is rendered
- **THEN** it SHALL display "Classic Mode" and "Sentence Mode" as options
- **AND** highlight the currently selected option
- **AND** show navigation hints ([↑↓] Select, [Enter] Confirm, [ESC] Back)

#### Scenario: Navigate mode selection
- **GIVEN** the mode selection screen is active
- **WHEN** user presses up or down arrow keys
- **THEN** the selection SHALL move to the previous or next mode
- **AND** update the highlighting accordingly

#### Scenario: Confirm mode selection
- **GIVEN** the mode selection screen is active
- **WHEN** user presses Enter
- **THEN** the system SHALL start the game in the selected mode
- **AND** transition to the game screen

### Requirement: Sentence Mode Game Rendering
The system SHALL render the sentence typing game interface.

#### Scenario: Display target sentence
- **GIVEN** the game is in Sentence mode and running
- **WHEN** the game screen is rendered
- **THEN** it SHALL display the complete target sentence
- **AND** use a distinct style (e.g., gray or muted color)

#### Scenario: Display user input with color coding
- **GIVEN** the game is in Sentence mode and user has typed characters
- **WHEN** the game screen is rendered
- **THEN** each typed character SHALL be displayed below or alongside the target
- **AND** correct characters SHALL be rendered in green
- **AND** incorrect characters SHALL be rendered in red
- **AND** untyped positions SHALL remain empty or show placeholders

#### Scenario: Show real-time statistics
- **GIVEN** the game is in Sentence mode
- **WHEN** the game screen is rendered
- **THEN** it SHALL display total characters typed
- **AND** display correct character count
- **AND** display current accuracy percentage
- **AND** display elapsed time

### Requirement: Sentence Mode Results Rendering
The system SHALL render completion results for Sentence mode.

#### Scenario: Display sentence mode results
- **GIVEN** the game in Sentence mode has finished
- **WHEN** the results screen is rendered
- **THEN** it SHALL show the target sentence
- **AND** show the user's typed sentence with color coding
- **AND** display total keystrokes
- **AND** display correct characters
- **AND** display accuracy percentage
- **AND** display typing speed (characters per second)
- **AND** display elapsed time
