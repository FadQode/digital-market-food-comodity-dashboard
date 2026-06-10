# Skills Directory

This directory contains specialized agent skill definitions for the Datathon Dicoding project.

## Purpose

Skills are specialized agent prompts that provide domain-specific expertise and guidelines for particular tasks. Each skill file contains detailed instructions, best practices, code patterns, and architectural guidance for a specific domain.

## Available Skills

### SCRAPER_AGENT.md
Expert Go web scraping agent specializing in building production-ready, scalable scraping solutions. This skill provides comprehensive guidance for:
- API client development (Apify, ScraperAPI, Bright Data)
- Direct HTML scraping with Go libraries
- Concurrent scraping with worker pools
- Error handling and resilience patterns
- Rate limiting and proxy management
- Data extraction and validation

## Usage

Reference these skill files when working on domain-specific tasks. The agent system can load these skills to provide specialized expertise and ensure consistent implementation patterns across the project.

## Adding New Skills

When adding new specialized agent skills:
1. Create a new `.md` file in this directory
2. Follow the existing pattern with clear sections for responsibilities, code examples, and best practices
3. Update this README with a brief description of the new skill
4. Ensure the skill aligns with the project's architecture and coding standards
