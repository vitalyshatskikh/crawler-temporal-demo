
import temporalio.client

from surfer.application import config, consts, workflows


async def setup_surfing(
    client: temporalio.client.Client,
    surf_conf: config.SurferConfig,
    schedules_conf: dict[str, str],
) -> None:
    for name, cron in schedules_conf.items():
        sc_id = f"{consts.WorkflowName.SEARCH_ADVERTS}/{name}"

        action_in = workflows.SearchAdvertsIn(
            surfer_config=surf_conf,
            surf_config_name=name,
        )

        action = temporalio.client.ScheduleActionStartWorkflow(
            workflows.SearchAdverts.run,
            action_in,
            id=sc_id,
            task_queue=consts.QueueName.SURFING_TASK,
            execution_timeout=surf_conf.search_adverts_timeout,
        )

        existing_schedule = client.get_schedule_handle(sc_id)
        await existing_schedule.delete()

        await client.create_schedule(
            id=sc_id,
            schedule=temporalio.client.Schedule(
                action=action,
                spec=temporalio.client.ScheduleSpec(cron_expressions=[cron]),
                policy=temporalio.client.SchedulePolicy(
                    overlap=temporalio.client.ScheduleOverlapPolicy.SKIP,
                ),
            ),
        )
