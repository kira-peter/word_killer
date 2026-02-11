# game-modes Specification

## Purpose
TBD - created by archiving change add-time-challenge-modes. Update Purpose after archive.
## Requirements
### Requirement: Time Challenge Game Modes
The system SHALL support three time-constrained game modes to provide diverse challenge types.

#### Scenario: Countdown mode definition
- **GIVEN** the game package is imported
- **WHEN** accessing the GameMode type
- **THEN** it SHALL provide a constant for Countdown mode (ModeCountdown)
- **AND** Countdown mode SHALL be distinct from Classic and Sentence modes

#### Scenario: Speed Run mode definition
- **GIVEN** the game package is imported
- **WHEN** accessing the GameMode type
- **THEN** it SHALL provide a constant for Speed Run mode (ModeSpeedRun)
- **AND** Speed Run mode SHALL be distinct from other modes

#### Scenario: Rhythm Master mode definition
- **GIVEN** the game package is imported
- **WHEN** accessing the GameMode type
- **THEN** it SHALL provide a constant for Rhythm Master mode (ModeRhythmMaster)
- **AND** Rhythm Master mode SHALL be distinct from other modes

### Requirement: Countdown Mode Initialization
The system SHALL initialize and manage Countdown mode with a 60-second time limit.

#### Scenario: Initialize countdown mode
- **GIVEN** a game instance and word dictionaries are loaded
- **WHEN** `StartCountdownMode(duration)` is called with 60 seconds
- **THEN** the game SHALL initialize in Countdown mode
- **AND** set the countdown duration to 60 seconds
- **AND** record the start time
- **AND** generate initial 30 words from multi-difficulty pools
- **AND** reset all game statistics

#### Scenario: Handle unloaded dictionaries
- **GIVEN** a game instance without loaded word dictionaries
- **WHEN** `StartCountdownMode(duration)` is called
- **THEN** the system SHALL return an error
- **AND** provide a descriptive error message about missing dictionaries

### Requirement: Countdown Mode Time Management
The system SHALL track remaining time and automatically end the game when time expires.

#### Scenario: Track countdown timer
- **GIVEN** the game is running in Countdown mode
- **WHEN** the elapsed time is calculated
- **THEN** the system SHALL compute remaining time as (duration - elapsed)
- **AND** update the timer every 100ms tick

#### Scenario: Display time warning
- **GIVEN** the game is running in Countdown mode
- **WHEN** remaining time is less than 10 seconds
- **THEN** the UI SHALL display the timer in red color with dark red background
- **AND** provide visual warning to the user

#### Scenario: Auto-end on timeout
- **GIVEN** the game is running in Countdown mode
- **WHEN** elapsed time reaches or exceeds the countdown duration
- **THEN** the game SHALL automatically transition to Finished status
- **AND** preserve all statistics up to that point

### Requirement: Countdown Mode Dynamic Word Generation
The system SHALL dynamically generate new words to maintain sufficient choices.

#### Scenario: Maintain minimum word count
- **GIVEN** the game is running in Countdown mode
- **WHEN** a word is eliminated and remaining words drop below 10
- **THEN** the system SHALL generate 20 new words from multi-difficulty pools
- **AND** append them to the current word list
- **AND** avoid duplicating already-completed words

#### Scenario: Continue word generation
- **GIVEN** the game is running in Countdown mode
- **WHEN** multiple words are eliminated rapidly
- **THEN** the system SHALL continue generating words as needed
- **AND** maintain the minimum threshold of 10 remaining words

### Requirement: Speed Run Mode Initialization
The system SHALL initialize Speed Run mode with a fixed set of 25 words.

#### Scenario: Initialize speed run mode
- **GIVEN** a game instance and word dictionaries are loaded
- **WHEN** `StartSpeedRunMode(targetWords)` is called with 25 words
- **THEN** the game SHALL initialize in Speed Run mode
- **AND** set the target word count to 25
- **AND** record the start time with nanosecond precision
- **AND** generate exactly 25 words from multi-difficulty pools
- **AND** reset all game statistics

#### Scenario: Handle unloaded dictionaries
- **GIVEN** a game instance without loaded word dictionaries
- **WHEN** `StartSpeedRunMode(targetWords)` is called
- **THEN** the system SHALL return an error
- **AND** provide a descriptive error message about missing dictionaries

### Requirement: Speed Run Mode Time Tracking
The system SHALL track elapsed time with millisecond precision for competitive timing.

#### Scenario: Track elapsed time
- **GIVEN** the game is running in Speed Run mode
- **WHEN** the elapsed time is calculated
- **THEN** the system SHALL compute time since start with millisecond precision
- **AND** format time as MM:SS.mmm (minutes:seconds.milliseconds)

#### Scenario: Auto-end on completion
- **GIVEN** the game is running in Speed Run mode
- **WHEN** all 25 words are eliminated
- **THEN** the game SHALL automatically transition to Finished status
- **AND** record the final completion time

### Requirement: Speed Run Mode Best Record Persistence
The system SHALL persist and load best completion times for Speed Run mode.

#### Scenario: Load existing best time
- **GIVEN** a `speedrun_record.json` file exists
- **WHEN** `loadSpeedRunBestTime()` is called
- **THEN** the system SHALL read the JSON file
- **AND** parse the best_time field (float64 seconds)
- **AND** return the best time value

#### Scenario: Handle missing record file
- **GIVEN** no `speedrun_record.json` file exists
- **WHEN** `loadSpeedRunBestTime()` is called
- **THEN** the system SHALL return 0
- **AND** NOT raise an error

#### Scenario: Save new best time
- **GIVEN** the game has finished in Speed Run mode
- **WHEN** the completion time is better than the current best time (or no record exists)
- **THEN** the system SHALL call `saveSpeedRunBestTime(newTime)`
- **AND** write a JSON file with the new best_time
- **AND** update the in-memory best time

#### Scenario: Preserve existing record
- **GIVEN** the game has finished in Speed Run mode
- **WHEN** the completion time is slower than the current best time
- **THEN** the system SHALL NOT update the record file
- **AND** keep the existing best time

### Requirement: Rhythm Master Mode Initialization
The system SHALL initialize Rhythm Master mode with per-word time limits.

#### Scenario: Initialize rhythm master mode
- **GIVEN** a game instance and word dictionaries are loaded
- **WHEN** `StartRhythmMasterMode()` is called
- **THEN** the game SHALL initialize in Rhythm Master mode
- **AND** set the initial word time limit to 2.0 seconds
- **AND** reset consecutive successes counter to 0
- **AND** reset difficulty level to 0
- **AND** generate initial 50 words from multi-difficulty pools
- **AND** record the start time for the first word
- **AND** reset all game statistics

#### Scenario: Handle unloaded dictionaries
- **GIVEN** a game instance without loaded word dictionaries
- **WHEN** `StartRhythmMasterMode()` is called
- **THEN** the system SHALL return an error
- **AND** provide a descriptive error message about missing dictionaries

### Requirement: Rhythm Master Mode Time Management
The system SHALL track per-word time limits and immediately fail on timeout.

#### Scenario: Track current word timer
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** the current word start time is recorded
- **THEN** the system SHALL calculate elapsed time for the current word
- **AND** display remaining time with a progress bar (green → yellow → red)

#### Scenario: Auto-fail on word timeout
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** elapsed time for the current word reaches or exceeds the time limit
- **THEN** the game SHALL immediately transition to Finished status with Aborted flag
- **AND** display Game Over message

#### Scenario: Check timeout on key press
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** the user presses a key via `AddChar()`
- **THEN** the system SHALL check if the current word has timed out before processing the input
- **AND** fail immediately if timeout occurred

#### Scenario: Check timeout on tick
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** the 100ms tick calls `CheckTimeouts()`
- **THEN** the system SHALL check if the current word has timed out
- **AND** fail immediately if timeout occurred
- **AND** allow the game to continue if time remains

### Requirement: Rhythm Master Mode Difficulty Progression
The system SHALL increase difficulty by reducing per-word time limits as the player progresses.

#### Scenario: Increment difficulty every 10 words
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** a word is successfully eliminated
- **THEN** the system SHALL increment the consecutive successes counter
- **AND** check if the counter is a multiple of 10
- **WHEN** the counter reaches a multiple of 10 (10, 20, 30, ...)
- **THEN** the system SHALL increment the difficulty level
- **AND** reduce the word time limit by 0.1 seconds

#### Scenario: Enforce minimum time limit
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** the calculated time limit is less than 0.5 seconds
- **THEN** the system SHALL clamp the time limit to 0.5 seconds
- **AND** NOT reduce it further

#### Scenario: Reset timer for next word
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** a word is successfully eliminated
- **THEN** the system SHALL find the next uncompleted word
- **AND** record the current time as the start time for that word
- **AND** apply the current difficulty's time limit

### Requirement: Rhythm Master Mode Dynamic Word Generation
The system SHALL dynamically generate new words to maintain infinite gameplay.

#### Scenario: Maintain minimum word count
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** a word is eliminated and remaining words drop below 10
- **THEN** the system SHALL generate 20 new words from multi-difficulty pools
- **AND** append them to the current word list
- **AND** avoid duplicating already-completed words

### Requirement: Time Check Integration
The system SHALL perform dual time checking for time-constrained modes.

#### Scenario: Check time on key press
- **GIVEN** the game is running in a time-constrained mode (Countdown or Rhythm Master)
- **WHEN** `AddChar()` is called
- **THEN** the system SHALL check mode-specific timeout conditions before processing input
- **AND** end the game immediately if timeout occurred

#### Scenario: Check time on tick
- **GIVEN** the game is running in a time-constrained mode (Countdown or Rhythm Master)
- **WHEN** the 100ms tick calls `CheckTimeouts()`
- **THEN** the system SHALL check mode-specific timeout conditions
- **AND** end the game immediately if timeout occurred

#### Scenario: Skip checks when not running
- **GIVEN** the game is NOT in Running status (Paused, Finished, etc.)
- **WHEN** `CheckTimeouts()` is called
- **THEN** the system SHALL immediately return without checking
- **AND** NOT modify game state

### Requirement: Mode-Specific Completion Logic
The system SHALL execute mode-specific logic when words are eliminated.

#### Scenario: Countdown mode word elimination
- **GIVEN** the game is running in Countdown mode
- **WHEN** a word is successfully eliminated via `TryEliminate()`
- **THEN** the system SHALL check if remaining words < 10
- **AND** generate 20 new words if threshold is met
- **AND** continue the game without auto-ending

#### Scenario: Speed Run mode word elimination
- **GIVEN** the game is running in Speed Run mode
- **WHEN** a word is successfully eliminated via `TryEliminate()`
- **THEN** the system SHALL check if all words are completed
- **AND** automatically end the game if all 25 words are eliminated
- **AND** record the final completion time

#### Scenario: Rhythm Master mode word elimination
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** a word is successfully eliminated via `TryEliminate()`
- **THEN** the system SHALL increment consecutive successes
- **AND** check and update difficulty level if needed
- **AND** reset the timer for the next word
- **AND** generate new words if remaining words < 10

### Requirement: Statistics Reuse
The system SHALL fully reuse the existing statistics system for all time challenge modes.

#### Scenario: Track statistics in Countdown mode
- **GIVEN** the game is running in Countdown mode
- **WHEN** keys are pressed and words are eliminated
- **THEN** the system SHALL use the existing Statistics methods
- **AND** track total keystrokes, valid keystrokes, correct characters, words completed, letters per second

#### Scenario: Track statistics in Speed Run mode
- **GIVEN** the game is running in Speed Run mode
- **WHEN** keys are pressed and words are eliminated
- **THEN** the system SHALL use the existing Statistics methods
- **AND** track elapsed time with millisecond precision
- **AND** calculate words per second and letters per second

#### Scenario: Track statistics in Rhythm Master mode
- **GIVEN** the game is running in Rhythm Master mode
- **WHEN** keys are pressed and words are eliminated
- **THEN** the system SHALL use the existing Statistics methods
- **AND** track letters per second, accuracy, and words completed
- **AND** additionally track consecutive successes in Game struct (not Statistics)

