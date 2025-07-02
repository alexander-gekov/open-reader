import { prisma } from "~/lib/prisma";

export default defineEventHandler(async (event) => {
  const body = await readBody(event);

  if (!body.email) {
    throw createError({
      statusCode: 400,
      message: "Email is required",
    });
  }

  try {
    const waitlistEntry = await prisma.waitlistEntry.create({
      data: {
        email: body.email,
        name: body.name,
      },
    });

    return {
      success: true,
      message: "Successfully joined waitlist",
      data: waitlistEntry,
    };
  } catch (error: any) {
    if (error.code === "P2002") {
      throw createError({
        statusCode: 400,
        message: "You're already on the waitlist!",
      });
    }

    console.error("Error creating waitlist entry:", error);
    throw createError({
      statusCode: 500,
      message: "Failed to join waitlist",
    });
  }
});
