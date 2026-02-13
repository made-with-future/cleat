# Specification: Ruby and Ruby on Rails Support

## Overview
This track implements first-class support for Ruby and the Ruby on Rails framework. Similar to the existing Python/Django implementation, Cleat will auto-detect Ruby projects via `Gemfile` and Rails projects via the presence of `bin/rails` and other markers. It will provide standardized commands for building (assets), running (server), and managing (migrations) Rails applications, with support for both local environments and Docker.

## Functional Requirements
- **Auto-Detection (`internal/detector`):**
    - Detect Ruby projects by searching for a `Gemfile` in the project root or service directories.
    - Detect Ruby on Rails projects by checking for `bin/rails` and `config/application.rb`.
    - Detect local Ruby environment managers: `rbenv` (`.ruby-version`), `rvm` (`.ruby-version` or `.rvmrc`), or `asdf` (`.tool-versions`).
- **Standardized Commands (`internal/strategy`):**
    - **`cleat build`**: Map to `bundle exec rails assets:precompile` (if Rails).
    - **`cleat run`**: Map to `bundle exec rails server`.
    - **`cleat ruby migrate`**: Map to `bundle exec rails db:migrate`.
- **Execution Logic (`internal/task`):**
    - **Docker Support**: If a Docker Compose service is identified for the Ruby module, run commands via `docker compose exec`.
    - **Local Support**: If no Docker is used, execute commands via the detected local environment (e.g., using `bundle exec`).
- **Configuration Schema (`internal/config/schema`):**
    - Add a `RubyConfig` struct to `ModuleConfig`.
    - Support options for specifying the Rails service name and custom entry points.

## Non-Functional Requirements
- **Consistency**: The user experience for Rails should be nearly identical to the Django experience.
- **Robustness**: Gracefully handle projects where `bundle` or `rails` are missing with clear error messages.

## Acceptance Criteria
- [ ] A project with a `Gemfile` is correctly detected as a Ruby project.
- [ ] A project with `bin/rails` is correctly detected as a Rails project.
- [ ] Running `cleat build` on a Rails project executes asset precompilation.
- [ ] Running `cleat run` on a Rails project starts the development server.
- [ ] Migrations can be triggered via a dedicated Cleat command.
- [ ] Commands prioritize Docker execution when configured.

## Out of Scope
- Support for Ruby frameworks other than Rails (e.g., Sinatra, Hanami) in this initial track.
- Automated installation of Ruby or environment managers.
