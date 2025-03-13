-- https://gitlab.com/gitlab-org/gitlab/-/blob/master/app/presenters/commit_status_presenter.rb
ALTER TABLE jobs
UPDATE failure_reason = CASE
    WHEN startsWith(failure_reason, 'There is an unknown failure') THEN 'unknown_failure'
    WHEN startsWith(failure_reason, 'There has been a script failure') THEN 'script_failure'
    WHEN (status = 'failed' AND failure_reason = '') THEN 'script_failure'
    WHEN startsWith(failure_reason, 'There has been an API failure') THEN 'api_failure'
    WHEN startsWith(failure_reason, 'There has been a timeout failure or the job got stuck') THEN 'stuck_or_timeout_failure'
    WHEN startsWith(failure_reason, 'There has been a runner system failure') THEN 'runner_system_failure'
    WHEN startsWith(failure_reason, 'There has been a missing dependency failure') THEN 'missing_dependency_failure'
    WHEN startsWith(failure_reason, 'No runners support the requirements to run this job') THEN 'runner_unsupported'
    WHEN startsWith(failure_reason, 'Your runner is outdated') THEN 'runner_unsupported'
    WHEN startsWith(failure_reason, 'Scheduled job could not be executed by some reason') THEN 'schedule_expired'
    WHEN startsWith(failure_reason, 'Delayed job could not be executed by some reason') THEN 'stale_schedule'
    WHEN startsWith(failure_reason, 'The script exceeded the maximum execution time set for the job') THEN 'job_execution_timeout'
    WHEN startsWith(failure_reason, 'The job is archived and cannot be run') THEN 'archived_failure'
    WHEN startsWith(failure_reason, 'The job failed to complete prerequisite tasks') THEN 'unmet_prerequisites'
    WHEN startsWith(failure_reason, 'The scheduler failed to assign job to the runner') THEN 'scheduler_failure'
    WHEN startsWith(failure_reason, 'There has been a structural integrity problem detected') THEN 'data_integrity_failure'
    WHEN startsWith(failure_reason, 'There has been an unknown job problem') THEN 'data_integrity_failure'
    WHEN startsWith(failure_reason, 'The deployment job is older than the previously succeeded deployment job') THEN 'forward_deployment_failure'
    WHEN startsWith(failure_reason, 'This job could not be executed because it would create infinitely looping pipelines') THEN 'pipeline_loop_detected'
    WHEN startsWith(failure_reason, 'This job could not be executed because of insufficient permissions to track the upstream project') THEN 'insufficient_upstream_permissions'
    WHEN startsWith(failure_reason, 'This job could not be executed because upstream bridge project could not be found') THEN 'upstream_bridge_project_not_found'
    WHEN startsWith(failure_reason, 'This job could not be executed because downstream pipeline trigger definition is invalid') THEN 'invalid_bridge_trigger'
    WHEN startsWith(failure_reason, 'This job could not be executed because downstream bridge project could not be found') THEN 'downstream_bridge_project_not_found'
    WHEN startsWith(failure_reason, 'The environment this job is deploying to is protected') THEN 'protected_environment_failure'
    WHEN startsWith(failure_reason, 'This job could not be executed because of insufficient permissions to create a downstream pipeline') THEN 'insufficient_bridge_permissions'
    WHEN startsWith(failure_reason, 'This job belongs to a child pipeline and cannot create further child pipelines') THEN 'bridge_pipeline_is_child_pipeline'
    WHEN startsWith(failure_reason, 'The downstream pipeline could not be created') THEN 'downstream_pipeline_creation_failed'
    WHEN startsWith(failure_reason, 'The secrets provider can not be found') THEN 'secrets_provider_not_found'
    WHEN startsWith(failure_reason, 'Maximum child pipeline depth has been reached') THEN 'reached_max_descendant_pipelines_depth'
    WHEN startsWith(failure_reason, 'You reached the maximum depth of child pipelines') THEN 'reached_max_descendant_pipelines_depth'
    WHEN startsWith(failure_reason, 'The downstream pipeline tree is too large') THEN 'reached_max_pipeline_hierarchy_size'
    WHEN startsWith(failure_reason, 'The job belongs to a deleted project') THEN 'project_deleted'
    WHEN startsWith(failure_reason, 'The user who created this job is blocked') THEN 'user_blocked'
    WHEN startsWith(failure_reason, 'No more CI minutes available') THEN 'ci_quota_exceeded'
    WHEN startsWith(failure_reason, 'No more compute minutes available') THEN 'ci_quota_exceeded'
    WHEN startsWith(failure_reason, 'No matching runner available') THEN 'no_matching_runner'
    WHEN startsWith(failure_reason, 'The job log size limit was reached') THEN 'trace_size_exceeded'
    WHEN startsWith(failure_reason, 'The CI/CD is disabled for this project') THEN 'builds_disabled'
    WHEN startsWith(failure_reason, 'This job could not be executed because it would create an environment with an invalid parameter') THEN 'environment_creation_failure'
    WHEN startsWith(failure_reason, 'This deployment job was rejected') THEN 'deployment_rejected'
    WHEN startsWith(failure_reason, 'This job could not be executed because group IP address restrictions') THEN 'ip_restriction_failure'
    WHEN startsWith(failure_reason, 'The deployment job is older than the latest deployment') THEN 'failed_outdated_deployment_job'
    WHEN startsWith(failure_reason, 'Too many downstream pipelines triggered in the last minute') THEN 'reached_downstream_pipeline_trigger_rate_limit'
    ELSE failure_reason
END
WHERE status = 'failed'
