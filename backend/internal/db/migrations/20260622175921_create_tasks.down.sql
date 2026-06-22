-- Delete the community leader keys table first (since it points to tasks)
DROP TABLE IF EXISTS task_keys;

-- Delete the main tasks table
DROP TABLE IF EXISTS tasks;

-- Finally, delete the custom dropdown lists we created
DROP TYPE IF EXISTS task_financial_state;
DROP TYPE IF EXISTS volunteer_mode;
DROP TYPE IF EXISTS task_status;