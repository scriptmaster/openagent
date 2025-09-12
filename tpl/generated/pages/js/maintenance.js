
                        function toggleMigrationStart() {
                            const checkbox = document.getElementById('reset_migrations');
                            const migrationField = document.getElementById('migration_start');
                            
                            if (checkbox.checked) {
                                migrationField.disabled = true;
                                migrationField.value = '0';
                            } else {
                                migrationField.disabled = false;
                                migrationField.value = '{page.MigrationStart}';
                            }
                        }
                        
